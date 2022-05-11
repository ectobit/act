.PHONY: lint test testv testcov

lint:
	@golangci-lint run --exclude-use-default=false --enable-all \
		--disable exhaustivestruct \
		--disable varnamelen

test:
	@go test ./...

testv:
	@go test -v ./...

testcov:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
