lint: ## Run linter
	golangci-lint run --fix

test: test/unit test/integration ## Run tests

test/unit: ## Run unit tests
	go test ./...

test/integration: ## Run integration tests
	cd ./tests/integration && \
		docker run -d --name lowbot-integration-test-redis --rm -e ALLOW_EMPTY_PASSWORD=yes -p 6379:6379 bitnami/redis:latest && \
		go test ./...; \
		docker kill lowbot-integration-test-redis

