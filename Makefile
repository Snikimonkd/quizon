LOCAL_BIN:=$(CURDIR)/bin
POSTGRES_PASSWORD:=some_password
PG_DSN:=postgres://postgres:$(POSTGRES_PASSWORD)@localhost:5432/postgres?sslmode=disable
DOMAIN:=localhost

run:
	PG_DSN=$(PG_DSN) DOMAIN=$(DOMAIN) go run cmd/main.go

test:
	PG_DSN=$(PG_DSN) DOMAIN=$(DOMAIN) go test -race -v -cover -coverprofile=cover.out.tmp -covermode=atomic -coverpkg ./... ./...

test-cov:test
	cat cover.out.tmp | grep -v "mock.go" | grep -v "/testsupport/" | grep -v "/generated/" > cover.out || cp cover.out.tmp cover.out
	go tool cover -func=cover.out
	go tool cover -html=cover.out

jet:bin-deps
	@PATH=$(LOCAL_BIN):$(PATH) jet -dsn $(PG_DSN) -path=./internal/generated/ -schema=public

bin-deps:
	GOPROXY="proxy.golang.org" GOBIN=$(LOCAL_BIN) go install github.com/go-jet/jet/v2/cmd/jet@latest
	GOPROXY="proxy.golang.org" GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest
	GOPROXY="proxy.golang.org" GOBIN=$(LOCAL_BIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	GOPROXY="proxy.golang.org" GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

gen-api:bin-deps
	@PATH=$(LOCAL_BIN):$(PATH) oapi-codegen --config=openapi/config.yaml openapi/api.yaml

lint:bin-deps
	@PATH=$(LOCAL_BIN):$(PATH) golangci-lint run --config=.golangci.yml

migrate-up:
	@echo "🆙 database migrations"
	@PATH=$(LOCAL_BIN):$(PATH) goose -dir migrations postgres "$(PG_DSN)" up

migrate-down:
	@echo "↩️  revert migration"
	PATH=$(LOCAL_BIN):$(PATH) goose -dir migrations postgres "$(PG_DSN)" down

create-migration:
	PATH=$(LOCAL_BIN):$(PATH) goose -dir migrations create $(NAME) sql
