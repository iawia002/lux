FILES_TO_FMT      ?= $(shell find . -path ./vendor -prune -o -name '*.go' -print)

## help: Show makefile commands
.PHONY: help
help: Makefile
	@echo "---- Project: kkdai/youtube ----"
	@echo " Usage: make COMMAND"
	@echo
	@echo " Management Commands:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## build: Build project
.PHONY: build
build:
	goreleaser --rm-dist

## deps: Ensures fresh go.mod and go.sum
.PHONY: deps
deps:
	go mod tidy
	go mod verify

## lint: Run golangci-lint check
.PHONY: lint
lint:
	command -v golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	echo "golangci-lint checking..."
	golangci-lint run --deadline=30m --enable=misspell --enable=gosec --enable=gofmt --enable=goimports --enable=revive ./cmd/... ./...
	go vet ./...

## format: Formats Go code
.PHONY: format
format:
	@echo ">> formatting code"
	@gofmt -s -w $(FILES_TO_FMT)

## test-unit: Run all Youtube Go unit tests
.PHONY: test-unit
test-unit:
	go test -v -cover ./...

## test-integration: Run all Youtube Go integration tests
.PHONY: test-integration
test-integration:
	mkdir -p output
	rm -f output/*
	ARTIFACTS=output go test -race -covermode=atomic -coverprofile=coverage.out -tags=integration ./...

.PHONY: coverage.out
coverage.out:

## clean: Clean files and downloaded videos from builds during development
.PHONY: clean
clean:
	rm -rf dist *.mp4 *.mkv
