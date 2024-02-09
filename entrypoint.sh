#!/bin/sh

echo "Updating CA certificates."
update-ca-certificates
echo "Running CMD."
redis-server --protected-mode no --daemonize yes
exec "$@"
