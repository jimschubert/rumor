package server

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/jimschubert/rumor/gen/rumor/v1"
	"github.com/jimschubert/rumor/internal/store"
)

// RumorServer implements the gRPC server for the Rumor service
type RumorServer struct {
	pb.UnimplementedRumorServiceServer
	store store.Store
}

// New RumorServer with the given store
func New(s store.Store) *RumorServer {
	return &RumorServer{store: s}
}

// List resources, with optional pagination
func (s *RumorServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	page, pageSize := int(req.Page), int(req.PageSize)
	if page < 1 {
		page = 1
	}

	records, total, err := s.store.List(req.Resource, req.Filters, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}

	items := make([]*structpb.Struct, 0, len(records))
	for _, r := range records {
		st, err := structpb.NewStruct(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "encoding record: %v", err)
		}
		items = append(items, st)
	}

	return &pb.ListResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// Get a resource by id
func (s *RumorServer) Get(ctx context.Context, req *pb.GetRequest) (*structpb.Struct, error) {
	r, err := s.store.Get(req.Resource, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}
	return toStruct(r)
}

// Create a resource
func (s *RumorServer) Create(ctx context.Context, req *pb.CreateRequest) (*structpb.Struct, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	r, err := s.store.Create(req.Resource, req.Data.AsMap())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return toStruct(r)
}

// Update a resource by id
func (s *RumorServer) Update(ctx context.Context, req *pb.UpdateRequest) (*structpb.Struct, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	r, err := s.store.Update(req.Resource, req.Id, req.Data.AsMap())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}
	return toStruct(r)
}

// Patch a resource by id
func (s *RumorServer) Patch(ctx context.Context, req *pb.PatchRequest) (*structpb.Struct, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "data is required")
	}
	r, err := s.store.Patch(req.Resource, req.Id, req.Data.AsMap())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "%v", err)
	}
	return toStruct(r)
}

// Delete a resource by id
func (s *RumorServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := s.store.Delete(req.Resource, req.Id); err != nil && !strings.Contains(err.Error(), "not found") {
		// note that "not found" means the record was deleted, we don't error in this case (idempotency)
		return nil, status.Error(codes.NotFound, "unexpected error")
	}
	return &pb.DeleteResponse{
		Success: true,
		Message: fmt.Sprintf("deleted %s/%s", req.Resource, req.Id),
	}, nil
}

// toStruct converts a store.Record into the dynamic protobuf struct
func toStruct(r store.Record) (*structpb.Struct, error) {
	st, err := structpb.NewStruct(r)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encoding: %v", err)
	}
	return st, nil
}
