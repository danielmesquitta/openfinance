.PHONY: default dev install test docs create_migration db_ui lint update build_lambda copy_lambda clear_lambda zip_lambda

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
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-w -s" -o ./tmp/bootstrap ./cmd/lambda/main.go
copy_lambda:
	@cp .env ./tmp/.env && cp users.json ./tmp/users.json
clear_lambda:
	@rm ./tmp/.env && rm ./tmp/users.json && rm ./tmp/bootstrap
zip_lambda:
	@zip -j ./tmp/lambda.zip ./tmp/bootstrap ./tmp/.env ./tmp/users.json
lambda:
	@make build_lambda && make copy_lambda && make zip_lambda && make clear_lambda
