.PHONY: default dev auth install

default: dev

dev:
	@air
install:
	@go mod tidy
docs:
	@swag init -g ./cmd/server/main.go -o ./internal/app/http/docs
