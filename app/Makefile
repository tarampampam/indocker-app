#!/usr/bin/make
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh
LDFLAGS = "-s -w -X gh.tarampamp.am/indocker-app/app/internal/version.version=$(shell git rev-parse HEAD)"
DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)"

.DEFAULT_GOAL : help

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fake-web-dist: # is needed for the backend (embedding)
	test -d ./web/dist || ( mkdir ./web/dist && touch ./web/dist/index.html )

# App stuff

go-generate: fake-web-dist ## Generate app assets
	docker-compose run $(DC_RUN_ARGS) --no-deps app sh -c "go generate ./... && go generate -tags docs ./internal/cli"

go-build: go-generate ## Build app binary file(s)
	docker-compose run $(DC_RUN_ARGS) -e "CGO_ENABLED=0" --no-deps go go build -trimpath -ldflags $(LDFLAGS) ./cmd/mkcert
	docker-compose run $(DC_RUN_ARGS) -e "CGO_ENABLED=0" --no-deps go go build -trimpath -ldflags $(LDFLAGS) ./cmd/app
	./app --version

go-test: fake-web-dist ## Run app tests
	docker-compose run $(DC_RUN_ARGS) --no-deps go gotestsum --format testname -- -race -timeout 10s ./...

go-lint: fake-web-dist ## Lint the app sources
	docker-compose run --rm golint golangci-lint run

go-fmt: fake-web-dist ## Run source code formatting tools
	docker-compose run $(DC_RUN_ARGS) --no-deps go gofmt -s -w -d .
	docker-compose run $(DC_RUN_ARGS) --no-deps go goimports -d -w .
	docker-compose run $(DC_RUN_ARGS) --no-deps go go mod tidy

go-shell: ## Start shell inside app environment
	docker-compose run $(DC_RUN_ARGS) go sh

# Frontend stuff

node-install: ## Install frontend dependencies
	docker-compose run $(DC_RUN_ARGS) --no-deps node npm install

node-lint: ## Lint the frontend sources
	docker-compose run $(DC_RUN_ARGS) --no-deps node npm run lint

node-build: ## Build the frontend
	docker-compose run $(DC_RUN_ARGS) --no-deps node npm run build

node-shell: ## Start shell inside frontend environment
	docker-compose run $(DC_RUN_ARGS) node sh

# Overall stuff

test: go-lint go-test node-lint ## Run all tests

up: go-generate ## Start the app in the development mode
	docker-compose up --detach app monitor whoami
	@printf "\n   \e[30;42m %s \033[0m"     'HTTPS ⇒ https://monitor.indocker.app';
	@printf "\n   \e[30;42m %s \033[0m\n\n" ' Press CTRL+C to stop logs watching ';
	docker-compose logs -f app monitor

down: ## Stop the app
	docker-compose down --remove-orphans

restart: down up ## Restart all containers

clean: ## Make clean
	docker-compose down -v -t 1
	-rm -R ./app ./mkcert ./web/dist
