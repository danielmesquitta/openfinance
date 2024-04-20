.PHONY: default dev auth install

default: dev

dev:
	@go run cmd/job/main.go
install:
	@go mod tidy
