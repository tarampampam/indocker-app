#!/usr/bin/make
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh
LDFLAGS = "-s -w -X gh.tarampamp.am/indocker-app/daemon/internal/version.version=$(shell git rev-parse HEAD)"

DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"
APP_NAME = $(notdir $(CURDIR))

.DEFAULT_GOAL : help

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Daemon stuff

daemon-generate: ## Generate daemon assets
	docker-compose run $(DC_RUN_ARGS) --no-deps daemon sh -c "go generate ./... && go generate -tags docs ./docs ./internal/cli"

daemon-build: daemon-generate ## Build daemon binary file
	docker-compose run $(DC_RUN_ARGS) -e "CGO_ENABLED=0" --no-deps daemon go build -trimpath -ldflags $(LDFLAGS) ./cmd/app
	./daemon/app --version

daemon-test: ## Run daemon tests
	docker-compose run $(DC_RUN_ARGS) --no-deps daemon gotestsum --format testname -- -race -timeout 10s ./...

daemon-lint: ## Lint the daemon sources
	docker-compose run --rm golint golangci-lint run

daemon-fmt: ## Run source code formatting tools
	docker-compose run $(DC_RUN_ARGS) --no-deps daemon gofmt -s -w -d .
	docker-compose run $(DC_RUN_ARGS) --no-deps daemon goimports -d -w .
	docker-compose run $(DC_RUN_ARGS) --no-deps daemon go mod tidy

daemon-shell: ## Start shell inside daemon environment
	docker-compose run $(DC_RUN_ARGS) daemon sh

test: daemon-lint daemon-test ## Run all tests

up: daemon-generate ## Start the app in the development mode
	docker-compose up --detach daemon-web whoami
	@printf "\n   \e[30;42m %s \033[0m"     'HTTP  ⇒ http://daemon.indocker.app';
	@printf "\n   \e[30;42m %s \033[0m\n"   'HTTPS ⇒ https://daemon.indocker.app';
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'Press CTRL+C to stop logs watching ';
	docker-compose logs -f daemon-web

down: ## Stop the app
	docker-compose down --remove-orphans

restart: down up ## Restart all containers

clean: ## Make clean
	docker-compose down -v -t 1
	-docker rmi $(APP_NAME):local -f
	-rm -R ./webhook-tester ./web/node_modules ./web/dist
