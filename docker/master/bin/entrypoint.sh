#!/bin/bash
set -e

sudo crond  -d 8 

echo "tart $@"
# Hand off to the CMD
exec "$@"