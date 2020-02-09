#!/bin/bash
# Inspired from https://github.com/hhcordero/docker-jmeter-client
# Basically runs jmeter, assuming the PATH is set to point to JMeter bin-dir (see Dockerfile)
#
# This script expects the standdard JMeter command parameters.
#
killall java
set -e

echo "START Running Jmeter on `date`"
echo "JVM_ARGS=${JVM_ARGS}"
echo "jmeter args=$@"

# Keep entrypoint simple: we must pass the standard JMeter arguments
#jmeter $@ upload/upload.jmx

#jmeter -n -N 127.0.0.1 upload/upload.jmx
nohup /opt/apache-jmeter/bin/jmeter-server -Dserver.rmi.localport=50000 -Dserver_port=1099  &
