.PHONY: up
up: ## Start the infrastructure in the background
	docker compose up

.PHONY: down
down: ## Stop all containers
	docker compose down

.PHONY: logs
logs: ## Tail container logs
	docker compose logs -f

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'