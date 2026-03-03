# Makefile for ensuring protoc and protoc-gen-grpc-gateway are generated consistently.
# Also provides build and test targets for the project for those familiar with Makefile.
# The fake-demo target creates a functional locally running demo.
PROTO_SRC = proto/rumor/v1/server.proto
OUT       = gen/

.PHONY: gen gen-verify build build-fake test fake-demo

gen: gen-verify $(PROTO_SRC)
	@mkdir -p $(OUT)
	protoc \
	  -I proto \
	  --go_out=$(OUT) --go_opt=paths=source_relative \
	  --go-grpc_out=$(OUT) --go-grpc_opt=paths=source_relative \
	  --grpc-gateway_out=$(OUT) --grpc-gateway_opt=paths=source_relative \
	  $(PROTO_SRC)

gen-verify:
	@command -v protoc >/dev/null 2>&1 || \
		(printf 'protoc is required to generate gRPC sources.\nInstall it from https://github.com/protocolbuffers/protobuf/releases or via `brew install protobuf`.\n'; exit 1)
	@command -v protoc-gen-grpc-gateway >/dev/null 2>&1 || \
		go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

build:
	@[ -f gen/rumor/v1/server.pb.go ] || $(MAKE) gen
	go build -o dist/rumor ./cmd/rumor

build-fake:
	go build -o dist/rumor-fake ./cmd/rumor-fake

test:
	go test ./...

fake-demo: build build-fake
	./dist/rumor-fake -c 50 users.schema.json -t -o /tmp/rumor-fake-demo.json
	./dist/rumor-fake -c 100 products.schema.json -t -o /tmp/rumor-fake-demo.json
	./dist/rumor-fake -c 20 addresses.schema.json -t -o /tmp/rumor-fake-demo.json
	@printf '\n**Starting rumor**\n\n' 2>&1
	./dist/rumor --db-path=/tmp/rumor-fake-demo.json
