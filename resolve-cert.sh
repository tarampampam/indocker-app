#!/usr/bin/env sh
set -e

echo "$0: Downloading"
curl -SsL -o /tmp/cert.tar.gz "$(echo aHR0cHM6Ly94LWNlcnQuaW5kb2NrZXIuYXBwL2FyY2hpdmUudGFyLmd6 | base64 -d)"

echo "$0: Extracting"
tar -C /tmp -xf /tmp/cert.tar.gz
rm /tmp/cert.tar.gz

echo "$0: Moving"
mv -f /tmp/fullchain.pem /etc/traefik/certs/fullchain1.pem
mv -f /tmp/privkey.pem /etc/traefik/certs/privkey1.pem

echo "$0: Setting permissions"
chmod 400 /etc/traefik/certs/*.pem
chown root:root /etc/traefik/certs/*.pem
