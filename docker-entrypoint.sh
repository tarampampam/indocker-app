#!/usr/bin/env sh
set -e

# in case of error do not stop the container, just log the error
/resolve-cert.sh || echo "$0: Failed to update TLS certificate"

exec "$@"
