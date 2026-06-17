.PHONY: run run-webhook run-backend test lint migrate-up migrate-down docker-up docker-down

run:
	$(MAKE) -C services/backend-service run

run-webhook:
	cd services/webhook-service && go run ./cmd

run-backend:
	$(MAKE) -C services/backend-service run

test:
	$(MAKE) -C services/backend-service test
	cd shared && go test ./... -count=1
	cd services/webhook-service && go test ./... -count=1

lint:
	$(MAKE) -C services/backend-service lint

migrate-up:
	$(MAKE) -C services/backend-service migrate-up

migrate-down:
	$(MAKE) -C services/backend-service migrate-down

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

clone:
	@echo "Usage: ./scripts/git-clone.sh <parent-repo-url> [dir]"

submodules:
	git submodule update --init --recursive

submodules-setup:
	./scripts/git-setup-submodules.sh
