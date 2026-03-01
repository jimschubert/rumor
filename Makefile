PROTO_SRC = proto/rumor/v1/server.proto
OUT       = gen/

.PHONY: gen build

gen:
	@command -v protoc-gen-grpc-gateway >/dev/null 2>&1 || \
		go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@mkdir -p $(OUT)
	protoc \
	  -I proto \
	  --go_out=$(OUT) --go_opt=paths=source_relative \
	  --go-grpc_out=$(OUT) --go-grpc_opt=paths=source_relative \
	  --grpc-gateway_out=$(OUT) --grpc-gateway_opt=paths=source_relative \
	  $(PROTO_SRC)

build: gen
	go build -o dist/rumor ./main.go
