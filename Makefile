LOCAL_BIN:=$(CURDIR)/bin
POSTGRES_PASSWORD:=quizon_db_password
LOCAL_DB_DSN:=postgres://postgres:$(POSTGRES_PASSWORD)@localhost:5432/postgres?sslmode=disable

run:
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) go run cmd/main.go

run-compose:
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) docker-compose -f ./docker-compose.yml up -d --no-deps --build --wait

stop-compose:
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) docker-compose -f ./docker-compose.yml up -d --no-deps --build --wait

test:
	go test -race -v -cover -coverprofile=cover.out.tmp -covermode=atomic -coverpkg ./... ./...

test-cov:test
	cat cover.out.tmp | grep -v "mock.go" | grep -v "/testsupport/" | grep -v "/generated/" > cover.out || cp cover.out.tmp cover.out
	go tool cover -func=cover.out
	go tool cover -html=cover.out

jet:bin-deps
	@PATH=$(LOCAL_BIN):$(PATH) jet -dsn $(LOCAL_DB_DSN) -path=./internal/generated/ -schema=public

bin-deps:
	GOPROXY="proxy.golang.org" GOBIN=$(LOCAL_BIN) go install github.com/go-jet/jet/v2/cmd/jet@latest
	GOPROXY="proxy.golang.org" GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest

lint:
	golangci-lint run --config=.golangci.yml

migrate-up:
	@echo "🆙 database migrations"
	@PATH=$(LOCAL_BIN):$(PATH) goose -dir migrations postgres "$(LOCAL_DB_DSN)" up

migrate-down:
	@echo "↩️  revert migration"
	PATH=$(LOCAL_BIN):$(PATH) goose -dir migrations postgres "$(LOCAL_DB_DSN)" down

create-migration:
	PATH=$(LOCAL_BIN):$(PATH) goose -dir migrations create $(NAME) sql
