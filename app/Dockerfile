# syntax=docker/dockerfile:1.3

# this stage is used to build the frontend
FROM docker.io/library/node:19-alpine as frontend

RUN mkdir -p /src

COPY ./web/package*.json /src/

WORKDIR /src

# install node dependencies
RUN set -x \
    && npm config set update-notifier false \
    && npm ci --no-audit --prefer-offline

COPY ./web /src

# build the frontend (built artifact can be found in /src/dist)
RUN set -x \
    && npm run build

# this stage is used to build the application
FROM docker.io/library/golang:1.20-alpine as builder

RUN set -x \
    # install common dependencies
    && apk add --no-cache git \
      # the following dependencies are needed for `go test` to work
      gcc musl-dev \
    && GOBIN=/usr/local/bin go install golang.org/x/tools/cmd/goimports@latest \
    && GOBIN=/usr/local/bin go install gotest.tools/gotestsum@latest

COPY ./go.* /src/

WORKDIR /src

# burn the modules cache
RUN set -x \
    # cache go dependencies
    && go mod download

# this stage is used to compile the application
FROM builder as compiler

COPY . /src

WORKDIR /src

# can be passed with any prefix (like `v1.2.3@GITHASH`)
# e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@GITHASH" .`
ARG APP_VERSION="undefined@docker"

ENV \
  # arguments to pass on each go tool link invocation
  LDFLAGS="-s -w -X gh.tarampamp.am/indocker-app/app/internal/version.version=${APP_VERSION}"

# copy the frontend build artifact
COPY --from=frontend /src/dist /src/web/dist

RUN set -x \
    && go generate ./... \
    && CGO_ENABLED=0 go build -trimpath -ldflags "${LDFLAGS}" -o ./app ./cmd/app/ \
    && ./app --version

# prepare rootfs for runtime
RUN mkdir -p /tmp/rootfs

WORKDIR /tmp/rootfs

RUN set -x \
    && mkdir -p \
        ./etc/ssl \
        ./bin \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && cp -R /etc/ssl/certs ./etc/ssl/certs \
    && mv /src/app ./bin/app

# and this is the final stage
FROM scratch as runtime

ARG APP_VERSION="undefined@docker"

LABEL \
    # docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.authors="tarampampam" \
    org.opencontainers.image.title="indocker.app" \
    org.opencontainers.image.description="Domain names with valid SSL for your local docker containers" \
    org.opencontainers.image.url="https://github.com/tarampampam/indocker-app" \
    org.opencontainers.image.source="https://github.com/tarampampam/indocker-app" \
    org.opencontainers.version="$APP_VERSION"

# import compiled application
COPY --from=compiler /tmp/rootfs /

# use an unprivileged user
USER 10001:10001

ENV \
  # logging format
  LOG_FORMAT=json \
  # logging level
  LOG_LEVEL=info

## docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
#HEALTHCHECK --interval=10s --timeout=3s --start-period=1s CMD [ \
#    "/bin/app", "http-server", "healthcheck" \
#]

ENTRYPOINT ["/bin/app"]
CMD ["start"]
