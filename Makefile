.PHONY: dev build down logs test fmt

dev:
	docker compose up --build

build:
	docker compose build

down:
	docker compose down

logs:
	docker compose logs -f

# Test & format backend lewat container Go — tak butuh toolchain Go lokal.
test:
	docker run --rm -v "$(CURDIR)/apps/backend":/src -w /src golang:1.25-alpine \
		sh -c "gofmt -l . && go vet ./... && go test ./..."

fmt:
	docker run --rm -v "$(CURDIR)/apps/backend":/src -w /src golang:1.25-alpine gofmt -w .

# Jalankan MCP (Model Context Protocol) bridge via stdio
mcp:
	go run ./apps/backend/cmd/mcp
