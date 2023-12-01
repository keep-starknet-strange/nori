LDFLAGSSTRING +=-X main.GitCommit=$(GITCOMMIT)
LDFLAGSSTRING +=-X main.GitDate=$(GITDATE)
LDFLAGSSTRING +=-X main.GitVersion=$(GITVERSION)
LDFLAGS := -ldflags "$(LDFLAGSSTRING)"

nori:
	go build -v $(LDFLAGS) -o ./bin/nori ./cmd/nori
.PHONY: nori

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
