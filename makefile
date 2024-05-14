.PHONY: default dev auth install

default: dev

dev:
	@air
install:
	@go mod tidy
