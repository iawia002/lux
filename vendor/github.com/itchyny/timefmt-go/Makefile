GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

.PHONY: all
all: test

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint -set_exit_status ./...

$(GOBIN)/golint:
	cd && go get golang.org/x/lint/golint

.PHONY: clean
clean:
	go clean
