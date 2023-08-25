GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOBIN=$(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif
OSARCH=$(shell uname -m)
GOLANGCI_LINT_PATH=$(GOBIN)/golangci-lint

.PHONY: install-git-hooks
install-git-hooks:
	@echo "Installing git hooks..."
	pre-commit install --hook-type pre-commit
	pre-commit install --hook-type commit-msg

.PHONY: install-tfplugindocs
install-tfplugindocs:
	@echo "Installing github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs..."
	@go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.16.0

.PHONY: install-golangci-lint
install-golangci-lint:
	@echo "Installing github.com/golangci/golangci-lint..."
	@(test -f $(GOLANGCI_LINT_PATH) && echo "github.com/golangci/golangci-lint is already installed. Skipping...") || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.54.2

.PHONY: install-tools
install-tools: install-golangci-lint install-tfplugindocs

.PHONY: install
install: install-tools install-git-hooks

.PHONY: generate
generate: install-tfplugindocs
	go generate ./...
