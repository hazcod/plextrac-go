.PHONY: all build test lint fmt vet cover bench tidy clean

all: test build

build:
	go build -o plextrac ./cmd/plextrac

test:
	go test -coverprofile=coverage.out ./...

lint:
	golangci-lint run --config=.github/golangci.yml

fmt:
	gofmt -s -w .
	goimports -w -local github.com/hazcod/plextrac-go .

vet:
	go vet ./...

update:
	go get -u ./... && go mod tidy

cover: test
	go tool cover -html=coverage.out -o coverage.html

bench:
	go test -bench=. -benchmem ./...

tidy:
	go mod tidy

clean:
	rm -f plextrac coverage.out coverage.html
	rm -rf dist/
