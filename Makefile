GOLANGCI_LINT := golangci-lint

default: help

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; \
			printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9\/-]+:.*?##/ \
			{ printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } \
			/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

test-targets = test/unit
.PHONY: test $(test-targets)
test: $(test-targets)  ## Run all tests
test/unit:
	go test -race './...'

.PHONY: validate
validate: ## Validate all OSV files
	go run ./cmd/validate -config ./config/config.yaml

.PHONY: preprocess
preprocess: ## Preprocess repository before assigning IDs
	go run ./cmd/preprocess -config ./config/config.yaml
