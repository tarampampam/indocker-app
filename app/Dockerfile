# syntax=docker/dockerfile:1

# -✂- this stage is used to develop and build the application locally -------------------------------------------------
FROM docker.io/library/node:21-bookworm AS develop

# install Go using the official image
COPY --from=docker.io/library/golang:1.22-bookworm /usr/local/go /usr/local/go

ENV \
  # add Go to the PATH
  PATH="$PATH:/go/bin:/usr/local/go/bin" \
  # use the /var/tmp/go as the GOPATH to reuse the modules cache
  GOPATH="/var/tmp/go" \
  # disable npm update notifier
  NPM_CONFIG_UPDATE_NOTIFIER=false

# install development tools and dependencies
RUN set -x \
    # renovate: source=github-releases name=golangci/golangci-lint
    && GOLANGCI_LINT_VERSION="1.59.1" \
    && wget -O- -nv "https://cdn.jsdelivr.net/gh/golangci/golangci-lint@v${GOLANGCI_LINT_VERSION}/install.sh" \
      | sh -s -- -b /bin "v${GOLANGCI_LINT_VERSION}"

WORKDIR /src/web

# install node dependencies
RUN \
    --mount=type=bind,source=web/package.json,target=/src/web/package.json \
    --mount=type=bind,source=web/package-lock.json,target=/src/web/package-lock.json \
    set -x \
    && npm config set update-notifier false \
    && npm ci -dd --no-audit --prefer-offline

WORKDIR /src

# burn the Go modules cache
RUN \
    --mount=type=bind,source=go.mod,target=/src/go.mod \
    --mount=type=bind,source=go.sum,target=/src/go.sum \
    go mod download -x \
    && find "${GOPATH}" -type d -exec chmod 0777 {} \; \
    && find "${GOPATH}" -type f -exec chmod 0666 {} \;

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM develop AS compile

# can be passed with any prefix (like `v1.2.3@FOO`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@FOO" .`
ARG APP_VERSION="undefined@docker"

# copy the source code
COPY . .

# build the frontend (built artifact can be found in /src/dist)
RUN set -x \
    && cd ./web \
    && npm run build

RUN set -x \
    # build the app itself
    && go generate ./... \
    && CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X gh.tarampamp.am/indocker-app/app/internal/version.version=${APP_VERSION}" \
      -o ./app \
      ./cmd/app/ \
    && ./app --version \
    # prepare rootfs for runtime
    && mkdir -p /tmp/rootfs \
    && cd /tmp/rootfs \
    && mkdir -p ./etc/ssl/certs ./bin ./tmp \
    && echo 'appuser:x:10001:10001::/nonexistent:/sbin/nologin' > ./etc/passwd \
    && echo 'appuser:x:10001:' > ./etc/group \
    && chmod 777 ./tmp \
    && cp /etc/ssl/certs/ca-certificates.crt ./etc/ssl/certs/ \
    && mv /src/app ./bin/app

# -✂- and this is the final stage -------------------------------------------------------------------------------------
FROM scratch AS runtime

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
COPY --from=compile /tmp/rootfs /

# use an unprivileged user
USER 10001:10001

ENV \
  # logging format
  LOG_FORMAT=json \
  # logging level
  LOG_LEVEL=info

HEALTHCHECK --interval=10s --start-interval=1s --start-period=5s --timeout=1s CMD ["/bin/app", "http-server", "healthcheck"]
ENTRYPOINT ["/bin/app"]
CMD ["start"]
