.PHONY: lint test testv testcov

lint:
	@golangci-lint run --exclude-use-default=false --enable-all \
		--disable golint \
		--disable interfacer \
		--disable scopelint \
		--disable maligned \
		--disable varnamelen

test:
	@go test ./...

testv:
	@go test -v ./...

testcov:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
