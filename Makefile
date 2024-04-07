LOCAL_DB_DSN:=$(shell grep "pg-dsn" .config/local_config.yaml | tail -n1 | sed "s/.*pg-dsn: //g" | sed "s/\"//g")
START:=postgresql:\/\/
DATABASE_USER:=$(shell echo "$(LOCAL_DB_DSN)" | grep -oE "^${START}[^:]*" | sed "s/${START}//g" )
START:=${START}${DATABASE_USER}:
DATABASE_PASSWORD:=$(shell echo "$(LOCAL_DB_DSN)" | grep -oE "^${START}[^@]*" | sed "s/${START}//g" )
START:=${START}${DATABASE_PASSWORD}@
DATABASE_HOST:=$(shell echo "$(LOCAL_DB_DSN)" | grep -oE "^${START}[^:]*" | sed "s/${START}//g" )
START:=${START}${DATABASE_HOST}:
DATABASE_PORT:=$(shell echo "$(LOCAL_DB_DSN)" | grep -oE "^${START}[^/]*" | sed "s/${START}//g" )
START:=${START}${DATABASE_PORT}\/
DATABASE_NAME:=$(shell echo "$(LOCAL_DB_DSN)" | grep -oE "^${START}[^\?]*" | sed "s/${START}//g")

kek:
	echo "$(DATABASE_USER)"
	echo "$(DATABASE_PASSWORD)"
	echo "$(DATABASE_HOST)"
	echo "$(DATABASE_PORT)"
	echo "$(DATABASE_NAME)"
	echo "$(START)"

run:
	go run ./cmd/main.go

create-migration\:%:
	goose -dir migrations create $* sql

migrate-up:
	goose -dir migrations postgres "$(LOCAL_DB_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(LOCAL_DB_DSN)" down

postgres-up:
	docker run --name quizon -p $(DATABASE_PORT):5432 -e POSTGRES_USER=$(DATABASE_USER) -e POSTGRES_DB=$(DATABASE_NAME) -e POSTGRES_PASSWORD=$(DATABASE_PASSWORD) -d postgres
