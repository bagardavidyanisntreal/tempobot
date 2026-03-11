#!/bin/sh
set -e

mkdir -p /app/data
chmod 777 /app/data || true

exec "$@"
