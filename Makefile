lint: ## Run linter
	golangci-lint run --fix

test: ## Run test
	go test ./...
