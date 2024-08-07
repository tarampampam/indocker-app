# syntax=docker/dockerfile:1.3

FROM traefik:3.1 as builder

# prepare rootfs for runtime
WORKDIR /tmp/rootfs

RUN set -x \
    && mkdir -p \
      ./etc/ssl \
      ./etc/traefik/certs \
      ./opt/traefik \
      ./bin \
      ./tmp \
    && chmod 777 ./tmp \
    && cp -R /etc/ssl/certs ./etc/ssl/certs \
    && mv /usr/local/bin/traefik ./bin/traefik \
    && chmod 755 ./bin/traefik \
    && chown -c 0:0 ./bin/traefik

# install curl for healthcheck
COPY --from=tarampampam/curl:7.87.0 /bin/curl ./bin/curl

# create runtime image
FROM scratch as runtime

# import rootfs from builder
COPY --from=builder /tmp/rootfs /

# copy configs...
COPY ./traefik/configs/traefik.yaml /etc/traefik/traefik.yaml
COPY ./traefik/configs/dynamic /etc/traefik/dynamic

# ...certs...
COPY ./traefik/certs/*.pem /etc/traefik/certs/

# ...and plugins
COPY ./traefik/plugins /opt/traefik/plugins-local/src

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

# docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
HEALTHCHECK --interval=5s --timeout=3s --start-period=1s CMD [ \
    "/bin/curl", "--fail", "http://127.0.0.1:81/ping" \
]

ENTRYPOINT ["/bin/traefik"]
