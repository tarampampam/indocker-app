#!/usr/bin/make
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh
LDFLAGS = "-s -w -X gh.tarampamp.am/indocker-app/app/internal/version.version=$(shell git rev-parse HEAD)"

DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"
APP_NAME = $(notdir $(CURDIR))

.DEFAULT_GOAL : help

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# App stuff

app-generate: ## Generate app assets
	docker-compose run $(DC_RUN_ARGS) --no-deps app sh -c "go generate ./... && go generate -tags docs ./docs ./internal/cli"

app-build: app-generate ## Build app binary file
	docker-compose run $(DC_RUN_ARGS) -e "CGO_ENABLED=0" --no-deps app go build -trimpath -ldflags $(LDFLAGS) ./cmd/mkcert
	docker-compose run $(DC_RUN_ARGS) -e "CGO_ENABLED=0" --no-deps app go build -trimpath -ldflags $(LDFLAGS) ./cmd/app
	./app/app --version

app-test: ## Run app tests
	docker-compose run $(DC_RUN_ARGS) --no-deps app gotestsum --format testname -- -race -timeout 10s ./...

app-lint: ## Lint the app sources
	docker-compose run --rm golint golangci-lint run

app-fmt: ## Run source code formatting tools
	docker-compose run $(DC_RUN_ARGS) --no-deps app gofmt -s -w -d .
	docker-compose run $(DC_RUN_ARGS) --no-deps app goimports -d -w .
	docker-compose run $(DC_RUN_ARGS) --no-deps app go mod tidy

app-shell: ## Start shell inside app environment
	docker-compose run $(DC_RUN_ARGS) app sh

test: app-lint app-test ## Run all tests

up: app-generate ## Start the app in the development mode
	docker-compose up --detach app-web whoami
	@printf "\n   \e[30;42m %s \033[0m"     'HTTP  ⇒ http://app.indocker.app';
	@printf "\n   \e[30;42m %s \033[0m\n"   'HTTPS ⇒ https://app.indocker.app';
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'Press CTRL+C to stop logs watching ';
	docker-compose logs -f app-web

down: ## Stop the app
	docker-compose down --remove-orphans

restart: down up ## Restart all containers

clean: ## Make clean
	docker-compose down -v -t 1
	-docker rmi $(APP_NAME):local -f
	-rm -R ./webhook-tester ./web/node_modules ./web/dist
