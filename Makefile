.PHONY: lint test test-verbose test-with-coverage

lint:
	@golint ./...
	@golangci-lint run --enable-all

test:
	@go test ./...

testv:
	@go test -v ./...

testcov:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
