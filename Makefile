LDFLAGSSTRING +=-X main.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X main.GitDate=$(GITDATE)
LDFLAGSSTRING +=-X main.GitVersion=$(GITVERSION)
LDFLAGS := -ldflags "$(LDFLAGSSTRING)"

starknet-proxyd:
	go build -v $(LDFLAGS) -o ./bin/starknet-proxyd ./cmd/starknet-proxyd
.PHONY: starknet-proxyd

fmt:
	go mod tidy
	gofmt -w .
.PHONY: fmt

test:
	go test -v ./...
.PHONY: test

lint:
	go vet ./...
.PHONY: test
