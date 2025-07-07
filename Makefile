GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOBIN=$(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif
OSARCH=$(shell uname -m)
GOLANGCI_LINT_PATH=$(GOBIN)/golangci-lint
TFPLUGINDOCS_PATH=$(GOBIN)/tfplugindocs

.PHONY: install-git-hooks
install-git-hooks:
	@echo "Installing git hooks..."
	pre-commit install --hook-type pre-commit
	pre-commit install --hook-type commit-msg

.PHONY: install-tfplugindocs
install-tfplugindocs:
	@echo "Installing github.com/hashicorp/terraform-plugin-docs..."
	@(test -f $(TFPLUGINDOCS_PATH) && echo "github.com/hashicorp/terraform-plugin-docs is already installed. Skipping...") || (wget -q -O $(TFPLUGINDOCS_PATH).zip https://github.com/hashicorp/terraform-plugin-docs/releases/download/v0.22.0/tfplugindocs_0.22.0_$(GOOS)_$(GOARCH).zip && unzip $(TFPLUGINDOCS_PATH).zip tfplugindocs -d $(GOBIN) && rm $(TFPLUGINDOCS_PATH).zip)

.PHONY: install-golangci-lint
install-golangci-lint:
	@echo "Installing github.com/golangci/golangci-lint..."
	@(test -f $(GOLANGCI_LINT_PATH) && echo "github.com/golangci/golangci-lint is already installed. Skipping...") || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v2.2.1

.PHONY: install-tools
install-tools: install-golangci-lint install-tfplugindocs

.PHONY: install
install: install-tools install-git-hooks

.PHONY: generate
generate: install-tfplugindocs
	go generate ./...
