#!/bin/bash
set -e

sudo crond  -d 8 

nohup /home/jmeter/jmeter-master/bin/jmeter-master &> jmeter-master.out &

echo "tart $@"
# Hand off to the CMD
exec "$@"