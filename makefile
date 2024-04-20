.PHONY: default dev auth install

default: dev

dev:
	@go run cmd/server/main.go
install:
	@go mod tidy
