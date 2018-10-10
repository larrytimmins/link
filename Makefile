test:
	go vet ./...
	go test -race `go list ./... | grep -v /vendor/`

test-verbose:
	go test -race -v `go list ./... | grep -v /vendor/`

fmt:
	godep go fmt

generate:
	go build -i ./cmd/generate-mock
	./generate-mock