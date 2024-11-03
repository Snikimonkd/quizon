LOCAL_DB_DSN:=$(shell grep -A1 'database' config/config.yaml | tail -n1 | sed "s/.*dsn: //g" | sed "s/\"//g")

LOCAL_BIN := $(CURDIR)/bin

run:
	ENV="local" go run cmd/main.go

run-compose:
	docker compose up build -d

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
