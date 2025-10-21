NAME=inngen
TS=$(shell date -u +"%FT%T")
_TAG := $(shell git tag | sort -V | tail -1)
TAG := $(if $(_TAG),$(_TAG),v0.0.0)
VERSION=$(shell git rev-parse --short HEAD)
LDFLAGS=-X main.Version=$(TAG) -X main.Revision=git:$(VERSION) -X main.BuildDate=$(TS)

.PHONY: all fmt lint test build run clean

all: build

precommit: lint test build

fmt:
	gofmt -d .

lint:
	@test -z "$$(gofmt -l .)" || (echo "ERROR: gofmt found formatting issues. Run 'make fmt'." && exit 1)
	@echo "gofmt successful"

	@-golangci-lint run -c ./golangci.yml $(PWD)/...
	@-govulncheck $(PWD)/...
	@-staticcheck $(PWD)/...
	@-gosec $(PWD)/...

test: lint
	# go test -v -race -cover -coverprofile=coverage.out -trace trace.out github.com/z0rr0/inngen
	# go tool cover -html=coverage.out
	go test -race -cover $(PWD)/...

build:
	go build -o $(NAME) -ldflags "$(LDFLAGS)" .

run: build
	./$(NAME) -w

#tools:
#	@go get -tool github.com/securego/gosec/v2/cmd/gosec@latest
#	@go get -tool honnef.co/go/tools/cmd/staticcheck@latest
#	@go get -tool golang.org/x/vuln/cmd/govulncheck@latest

clean:
	@rm -f $(NAME)
	@rm -rf $(COVER_DIR)/*

