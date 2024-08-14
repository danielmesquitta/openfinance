.PHONY: default dev install test docs create_migration db_ui lint update

default: dev

dev:
	@air
install:
	@go mod download
test:
	@go test -count=1 -v ./...
docs:
	@swag init -g ./cmd/server/main.go -o ./internal/app/http/docs
create_migration:
	@npx prisma migrate dev --create-only --schema=./sql/schema.prisma
migrate:
	@npx prisma migrate deploy --schema=./sql/schema.prisma
db_generate:
	@sqlc generate
db_ui:
	@npx prisma studio --schema=./sql/schema.prisma
lint:
	@golangci-lint run && nilaway ./...
update:
	@go mod tidy && go get -u ./...
build_lambda:
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-w -s" -o ./bin/bootstrap ./cmd/lambda/main.go
zip_lambda:
	@make build_lambda && cp .env ./bin/.env && zip -j ./bin/lambda.zip ./bin/bootstrap ./bin/.env && rm ./bin/.env
