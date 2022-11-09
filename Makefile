MODULE = $(shell go list -m)

.PHONY: generate build test lint build-docker compose compose-down migrate
generate:
	go generate ./...

build: # build a server
	go build -a -o kek-server $(MODULE)/cmd/server

start: # start a server
	./kek-server server

migrate: # migrate database
	migrate -path ./migrations -database "postgres://common:@localhost:5432/kek?sslmode=disable" up

force: # migrate force
	migrate -path ./migrations -database "postgres://common:@localhost:5432/kek?sslmode=disable" force 0

test:
	go clean -testcache
	go test ./... -v

lint:
	gofmt -l .

# build-docker: # build docker image
# 	docker build -f cmd/server/Dockerfile -t kek-server-v2/kek-server .

# compose: # run with docker-compose
# 	docker-compose up --force-recreate

# compose-down: # down docker-compose
# 	docker-compose down -v

# force: # force migration
# 	docker run --rm -v migrations:/migrations --network host migrate/migrate -path ./migrations \
# 	-database "host=localhost user=root password=password dbname=kek port=9920 sslmode=disable TimeZone=Asia/Tokyo" force 1

# migrate:
# 	docker run --rm -v migrations:/migrations --network host migrate/migrate -path ./migrations \
# 	-database "host=localhost user=root password=password dbname=kek port=9920 sslmode=disable TimeZone=Asia/Tokyo" up 2



# migrate create -ext sql -seq -dir ./migrations users
