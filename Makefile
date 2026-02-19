BINARY := conductor-powerline
MODULE := github.com/rbarcante/conductor-powerline

.PHONY: build test test-coverage lint install clean fmt vet

build:
	go build -o $(BINARY) .

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run

install:
	go install $(MODULE)

clean:
	rm -f $(BINARY) coverage.out coverage.html

fmt:
	go fmt ./...

vet:
	go vet ./...
