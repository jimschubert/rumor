# rumor

[![License](https://img.shields.io/github/license/jimschubert/rumor?color=blue)](./LICENSE)
![Go Version](https://img.shields.io/github/go-mod/go-version/jimschubert/rumor)

Lightweight server (`rumor`) for serving mock data via gRPC and HTTP/JSON APIs; backed by JSON files for storage.

Use `rumor-fake` to generate fake data from JSON-based schema (supports both simplified schema and standard JSON Schema formats).

## Install

**rumor**
```shell
go install github.com/jimschubert/rumor/cmd/rumor@latest
```
**rumor-fake**
```shell
go install github.com/jimschubert/rumor/cmd/rumor-fake@latest
```

Or build from source:

```shell
git clone https://github.com/jimschubert/rumor
cd rumor
make build build-fake
```

## Usage

Generate fake data using an embedded example schema (or bring your own):

```shell
./dist/rumor-fake users.schema.json
```

This creates `db.json` with 10 fake user records. The tool includes embedded examples (`users.schema.json`, 
`products.schema.json`, `addresses.schema.json`) that can be referenced by name, 
or you can provide a path to your own schema file. 

You can generate multiple schemas into the same existing file.
Run `./dist/rumor-fake --help` for more options.

Start the server:

```shell
./dist/rumor
```

Access your data at `http://localhost:8080/api/users` or via gRPC on port `9090`.

>[!NOTE]
> The gRPC server exposes reflection, so you can more easily discover usage via Postman (or other tools supporting gRPC reflection)

### Options: rumor

```shell
./dist/rumor --help
Usage: rumor [flags]

A simple gRPC/HTTP server for storing and retrieving JSON records, with a file-based JSON database.

Flags:
  -h, --help                             Show context-sensitive help.
      --db-path="db.json"                Path to JSON database file
      --grpc-address="localhost:9090"    gRPC TCP listen address
      --http-address="localhost:8080"    HTTP/JSON listen address
  -v, --version                          Print version information
```

Example:
```shell
./dist/rumor --db-path data.json --http-port 3000 --grpc-port 5000
```

### Options: rumor-fake

```shell
./dist/rumor-fake --help
Usage: rumor-fake <schema> [flags]

Generate fake data for the rumor server based on a schema definition.

Arguments:
  <schema>    Path to JSON schema file defining the data structure

Flags:
  -h, --help                Show context-sensitive help.
  -c, --count=10            Number of records to generate
  -r, --resource=STRING     Resource name (defaults to schema filename without extension)
  -o, --output="db.json"    Output JSON database file
  -v, --version             Print version information
```

Example:
```shell
./dist/rumor-fake -c 50 users.schema.json -o /tmp/rumor-fake-demo.json
```

## Schema Definition

Both simplified and standard JSON Schema formats are supported with automatic detection.

**Simplified format** (`users.schema.json`):

```json
{
  "fields": {
    "email": {"type": "email"},
    "first_name": {"type": "first_name"},
    "last_name": {"type": "last_name"},
    "company": {"type": "company"},
    "status": {"type": "string", "value": "active"}
  }
}
```

**JSON Schema format**:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "email": {"type": "string", "format": "email"},
    "first_name": {"type": "string", "format": "first_name"},
    "last_name": {"type": "string", "format": "last_name"},
    "company": {"type": "string", "format": "company"},
    "status": {"type": "string", "const": "active"}
  }
}
```

See [internal/faker/doc.go](internal/faker/doc.go) for supported field types and formats.

## Demo

You can try out a working demo with the provided example schemas:

```shell
make fake-demo
```
This will start the server:

* gRPC → tcp://localhost:9090
* REST → http://localhost:8080

## API

### HTTP/JSON

The HTTP endpoints are automatically generated based on the resource name.

For example, if your schema is `users.schema.json` (or you ran `make fake-demo` above), the resource is `users` and the endpoints are:

```shell
# GET /api/{resource}
curl http://localhost:8080/api/users

# GET /api/{resource}/{id}
curl http://localhost:8080/api/users/1

# POST /api/{resource}
# note: data does not have to be the same format for each user!
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","first_name":"Jane"}'

# PUT /api/{resource}/{id}
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{"email":"updated@example.com"}'

# DELETE /api/{resource}/{id}
curl -X DELETE http://localhost:8080/api/users/1
```

### gRPC

See the proto definitions in the repository for service contracts. Default gRPC port is `9090`.

## License

This project is [licensed](./LICENSE) under Apache 2.0.

