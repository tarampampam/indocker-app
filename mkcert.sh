#!/usr/bin/env sh
set -e

# Create a token (https://dash.cloudflare.com/profile/api-tokens) with the following permissions:
# - Zone:Zone:Read
# - Zone:DNS:Edit
# Zone Resources: Include -- Specific zone -- <your-root-domain>
CF_API_KEY="${CF_API_KEY:-}"

CF_EMAIL="${CF_EMAIL:-}" # email address for important account notifications
ROOT_DOMAIN="${ROOT_DOMAIN:-indocker.app}" # domain name to request a certificate for
DRY_RUN="${DRY_RUN:-true}" # true|false

if [ -n "$CF_EMAIL" ] && [ -n "$CF_API_KEY" ]; then
  server='https://acme-staging-v02.api.letsencrypt.org/directory'; # staging server

  if [ "$DRY_RUN" = "false" ]; then
    echo "$0: dry run mode disabled";
    server='https://acme-v02.api.letsencrypt.org/directory'; # production server
  fi;

  echo "$0: acme server: $server";
  echo "$0: generating certificate for $ROOT_DOMAIN";

  image='certbot/dns-cloudflare:v2.2.0'; # check for updates here - https://hub.docker.com/r/certbot/dns-cloudflare/tags

  docker pull --quiet "$image";
  docker run \
    -e "EMAIL=${CF_EMAIL}" \
    -e "API_KEY=${CF_API_KEY}" \
    -e "ROOT_DOMAIN=${ROOT_DOMAIN}" \
    -e "SERVER=$server" \
    -e "OUT_UID=$(id -u)" \
    -e "OUT_GID=$(id -g)" \
    -v "${PWD}/config/certs:/out:rw" \
    --entrypoint sh \
      "$image" -c \
        'echo -e "dns_cloudflare_api_token=${API_KEY}" > /tmp/credentials.ini \
        && set -x \
        && chmod 600 /tmp/credentials.ini \
        && certbot certonly \
          --non-interactive \
          --dns-cloudflare \
          --dns-cloudflare-credentials /tmp/credentials.ini \
          --dns-cloudflare-propagation-seconds 30 \
          --agree-tos \
          --domain "*.${ROOT_DOMAIN}" \
          --domain "*.app.${ROOT_DOMAIN}" \
          --domain "*.apps.${ROOT_DOMAIN}" \
          --domain "*.www.${ROOT_DOMAIN}" \
          --domain "*.http.${ROOT_DOMAIN}" \
          --domain "*.mail.${ROOT_DOMAIN}" \
          --domain "*.m.${ROOT_DOMAIN}" \
          --domain "*.go.${ROOT_DOMAIN}" \
          --domain "*.static.${ROOT_DOMAIN}" \
          --domain "*.img.${ROOT_DOMAIN}" \
          --domain "*.media.${ROOT_DOMAIN}" \
          --domain "*.admin.${ROOT_DOMAIN}" \
          --domain "*.api.${ROOT_DOMAIN}" \
          --domain "*.back.${ROOT_DOMAIN}" \
          --domain "*.backend.${ROOT_DOMAIN}" \
          --domain "*.front.${ROOT_DOMAIN}" \
          --domain "*.frontend.${ROOT_DOMAIN}" \
          --domain "*.srv.${ROOT_DOMAIN}" \
          --domain "*.service.${ROOT_DOMAIN}" \
          --domain "*.dev.${ROOT_DOMAIN}" \
          --domain "*.db.${ROOT_DOMAIN}" \
          --domain "*.test.${ROOT_DOMAIN}" \
          --domain "*.demo.${ROOT_DOMAIN}" \
          --domain "*.alpha.${ROOT_DOMAIN}" \
          --domain "*.beta.${ROOT_DOMAIN}" \
          --domain "*.x-docker.${ROOT_DOMAIN}" \
          --email "${EMAIL}" \
          --server "${SERVER}"; \
        rm -f /tmp/credentials.ini \
        && chown -R "${OUT_UID}:${OUT_GID}" /etc/letsencrypt/archive/${ROOT_DOMAIN} \
        && chmod 644 /etc/letsencrypt/archive/${ROOT_DOMAIN}/*.pem \
        && mv -fv /etc/letsencrypt/archive/${ROOT_DOMAIN}/*.pem /out';
else
  >&2 echo "$0: no CF credentials provided";
  exit 1;
fi;
