.PHONY: test build run migrate asyncapi docs help

test: ## Run all tests
	@go test -v -json ./... | gotestdox

build: ## Build server and client binaries
	@go build -o bin/server ./cmd/server
	@go build -o bin/takgo ./cmd/client

run: ## Run the server
	@go run ./cmd/server

migrate-up: ## Run database migrations
	@go run ./cmd/migrate up

migrate-down: ## Run database migrations
	@go run ./cmd/migrate down

asyncapi-generate: ## Generate asyncapi.yaml from Go types
	@go run ./cmd/asyncapi

async-docs: asyncapi ## Preview docs (uses asyncapi CLI if installed, else opens yaml)
	@asyncapi start preview asyncapi.yaml

help: ## Display this help message
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '(##@|##)' $(MAKEFILE_LIST) | grep -v grep | while read -r line; do \
		if [[ $$line =~ ^##@ ]]; then \
			echo ""; \
			echo "$${line####@ }"; \
		elif [[ $$line =~ ^[a-zA-Z_-]+: ]]; then \
			target=$$(echo "$$line" | cut -d':' -f1); \
			comment=$$(echo "$$line" | sed -n 's/.*## *//p'); \
			if [ -n "$$comment" ]; then \
				printf "    \033[32m%-20s\033[0m %s\n" "$$target" "$$comment"; \
			fi \
		fi \
	done
