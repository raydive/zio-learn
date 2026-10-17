GOLANGCI_LINT_VERSION := v2.15.0
GOLANGCI_LINT := $(CURDIR)/.bin/$(GOLANGCI_LINT_VERSION)/golangci-lint

.PHONY: run build test lint fmt fmt-check check tools

run:
	go run ./cmd/zio-learn

build:
	go build -trimpath -o .bin/zio-learn ./cmd/zio-learn

test:
	go test -race ./...

lint: fmt-check
	$(GOLANGCI_LINT) run ./...

fmt-check: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) fmt --diff

fmt: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) fmt

check: lint test build

tools: $(GOLANGCI_LINT)

$(GOLANGCI_LINT):
	mkdir -p .bin
	curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh -o .bin/install-golangci-lint.sh
	sh .bin/install-golangci-lint.sh -b $(dir $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION)
