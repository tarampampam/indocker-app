# syntax=docker/dockerfile:1

FROM docker.io/library/alpine:3.21 AS base

FROM base AS builder

WORKDIR /tmp/rootfs

COPY --from=docker.io/library/traefik:3.2 /usr/local/bin/traefik ./bin/traefik
COPY --from=ghcr.io/tarampampam/curl:8.11.1 /bin/curl ./bin/curl

RUN set -x \
    && mkdir -p \
      ./etc/traefik/certs \
      ./opt/traefik \
    && chown -c 0:0 ./bin/traefik

COPY ./traefik/configs/traefik.yaml ./etc/traefik/traefik.yaml
COPY ./traefik/configs/dynamic ./etc/traefik/dynamic
COPY ./docker-entrypoint.sh ./docker-entrypoint.sh

FROM base AS runtime

COPY --from=builder /tmp/rootfs /

ARG APP_VERSION="undefined"

LABEL \
    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.title="indocker.app" \
    org.opencontainers.image.description="Domain names with valid SSL for your local docker containers" \
    org.opencontainers.image.url="https://github.com/tarampampam/indocker-app" \
    org.opencontainers.image.source="https://github.com/tarampampam/indocker-app" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.version="$APP_VERSION" \
    org.opencontainers.image.licenses="MIT"

WORKDIR "/opt/traefik"

EXPOSE "80/tcp" "443/tcp"

HEALTHCHECK --interval=5s --start-interval=1s --start-period=5s CMD ["/bin/curl", "--fail", "http://127.0.0.1:81/ping"]

ENTRYPOINT ["/docker-entrypoint.sh", "/bin/traefik"]

#FROM docker.io/library/traefik:3.2 AS builder
#
## prepare the root fs for the runtime
#WORKDIR /tmp/rootfs
#
#RUN set -x \
#    && mkdir -p \
#      ./etc/ssl \
#      ./etc/traefik/certs \
#      ./opt/traefik \
#      ./bin \
#      ./tmp \
#    && chmod 777 ./tmp \
#    && cp -R /etc/ssl/certs ./etc/ssl/certs \
#    && mv /usr/local/bin/traefik ./bin/traefik \
#    && chmod 755 ./bin/traefik \
#    && chown -c 0:0 ./bin/traefik
#
## install curl for the healthcheck and certs resolving
#COPY --from=ghcr.io/tarampampam/curl:8.11.1 /bin/curl ./bin/curl
#
## create the runtime image
#FROM scratch AS runtime
#
## import rootfs from builder
#COPY --from=builder /tmp/rootfs /
#
## copy configs and plugins
#COPY ./traefik/configs/traefik.yaml /etc/traefik/traefik.yaml
#COPY ./traefik/configs/dynamic /etc/traefik/dynamic
#COPY ./traefik/plugins /opt/traefik/plugins-local/src
#
#ARG APP_VERSION="undefined"
#
#LABEL \
#    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
#    org.opencontainers.image.title="indocker.app" \
#    org.opencontainers.image.description="Domain names with valid SSL for your local docker containers" \
#    org.opencontainers.image.url="https://github.com/tarampampam/indocker-app" \
#    org.opencontainers.image.source="https://github.com/tarampampam/indocker-app" \
#    org.opencontainers.image.vendor="tarampampam" \
#    org.opencontainers.version="$APP_VERSION" \
#    org.opencontainers.image.licenses="MIT"
#
#WORKDIR "/opt/traefik"
#
#EXPOSE "80/tcp" "443/tcp"
#
## docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
#HEALTHCHECK --interval=5s --timeout=3s --start-period=1s CMD [ \
#    "/bin/curl", "--fail", "http://127.0.0.1:81/ping" \
#]
#
#ENTRYPOINT ["/bin/traefik"]
