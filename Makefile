.PHONY: lint test testv testcov

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
