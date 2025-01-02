#!/usr/bin/env sh
set -e

curl -SsL "$(echo aHR0cHM6Ly9pbmRvY2tlci1hcHAtY2VydHMucGFnZXMuZGV2L2FyY2hpdmUudGFyLmd6 | base64 -d)" | tar -xz -C /tmp
mv /tmp/fullchain.pem /etc/traefik/certs/fullchain1.pem
mv /tmp/privkey.pem /etc/traefik/certs/privkey1.pem
chmod 600 /etc/traefik/certs/*.pem

exec "$@"
