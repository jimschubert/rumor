package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jimschubert/rumor/internal/server"
	"github.com/jimschubert/rumor/internal/store"
	"github.com/jimschubert/rumor/internal/store/jsonstore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "github.com/jimschubert/rumor/gen/rumor/v1"
)

var (
	programName = "rumor"
	version     = "dev"
	commit      = "unknown SHA"
)

var CLI struct {
	DbPath      string           `default:"db.json" help:"Path to JSON database file"`
	GrpcAddress string           `default:"localhost:9090" help:"gRPC TCP listen address"`
	HttpAddress string           `default:"localhost:8080" help:"HTTP/JSON listen address"`
	Version     kong.VersionFlag `short:"v" help:"Print version information"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name(programName),
		kong.Description("A simple gRPC/HTTP server for storing and retrieving JSON records, with a file-based JSON database."),
		kong.UsageOnError(),
		kong.Vars{"version": fmt.Sprintf("%s (%s)", version, commit)},
	)

	if err := run(); err != nil {
		ctx.Errorf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	st, err := jsonstore.New(CLI.DbPath)
	if err != nil {
		return fmt.Errorf("store: %v", err)
	}

	errCh := make(chan error, 2)

	grpcShutdown := startGRPCServer(st, errCh)
	defer grpcShutdown()

	httpShutdown := startHTTPServer(errCh)
	defer httpShutdown()

	// Wait for shutdown signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var serverErr error
	select {
	case <-quit:
		log.Println("shutting down…")
	case serverErr = <-errCh:
		log.Printf("server error: %v", serverErr)
	}

	return serverErr
}

func startGRPCServer(st store.Store, errCh chan<- error) func() {
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			doNotPanicInterceptor,
		),
	)

	pb.RegisterRumorServiceServer(grpcSrv, server.New(st))
	reflection.Register(grpcSrv)

	tcpLis, err := net.Listen("tcp", CLI.GrpcAddress)
	if err != nil {
		errCh <- fmt.Errorf("gRPC TCP listen: %v", err)
		return func() {}
	}

	go func() {
		log.Printf("gRPC → tcp://%s", CLI.GrpcAddress)
		if err := grpcSrv.Serve(tcpLis); err != nil {
			errCh <- fmt.Errorf("gRPC TCP: %w", err)
		}
	}()

	return func() {
		grpcSrv.GracefulStop()
	}
}

func startHTTPServer(errCh chan<- error) func() {
	ctx, cancel := context.WithCancel(context.Background())

	gwMux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := pb.RegisterRumorServiceHandlerFromEndpoint(ctx, gwMux, CLI.GrpcAddress, dialOpts); err != nil {
		errCh <- fmt.Errorf("gateway register: %v", err)
		cancel()
		return func() {}
	}

	httpSrv := &http.Server{
		Addr:         CLI.HttpAddress,
		Handler:      gwMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		//goland:noinspection HttpUrlsUsage
		log.Printf("REST → http://%s", CLI.HttpAddress)
		if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("HTTP: %w", err)
		}
	}()

	return func() {
		cancel()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		if err := httpSrv.Shutdown(shutCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}
	}
}

var loggingInterceptor grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("%-50s %v", info.FullMethod, time.Since(start))
	return resp, err
}

var doNotPanicInterceptor grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in %s: %v", info.FullMethod, r)
			err = status.Errorf(500, "internal server error")
		}
	}()
	return handler(ctx, req)
}
