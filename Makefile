PROTO_SRC = proto/rumor/v1/server.proto
OUT       = gen/

.PHONY: gen build gen-verify

gen: gen-verify
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

build: gen
	go build -o dist/rumor ./main.go
