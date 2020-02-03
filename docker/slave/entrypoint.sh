#!/bin/bash
# Inspired from https://github.com/hhcordero/docker-jmeter-client
# Basically runs jmeter, assuming the PATH is set to point to JMeter bin-dir (see Dockerfile)
#
# This script expects the standdard JMeter command parameters.
#
set -e

echo "START Running Jmeter on `date`"
echo "JVM_ARGS=${JVM_ARGS}"
echo "jmeter args=$@"

# Keep entrypoint simple: we must pass the standard JMeter arguments
#jmeter $@ upload/upload.jmx

#jmeter -n -N 127.0.0.1 upload/upload.jmx
#/opt/apache-jmeter-5.1.1/bin/jmeter-server -Dserver.rmi.localport=50000 -Dserver_port=1099
#!/bin/bash

export TENNANT="$(cut -d'-' -f1 <<<"$2")"
TEST_RUNNING="$(pidof jmeter)"
echo "JMETER_PID=$TEST_RUNNING TENNAT=$TENNANT"
if [ -z "$TEST_RUNNING" ]; then
    cat >/start.sh <<EOF
#!/usr/bin/env bash

freeMem=`awk '/MemFree/ { print int($2/1024) }' /proc/meminfo`
s=$(($freeMem/10*8))
x=$(($freeMem/10*8))
n=$(($freeMem/10*2))
export JVM_ARGS="-Xmn${}m -Xms${s}m -Xmx${x}m"


echo 1
if test -f /tmp/start; then
echo 2
    PID=\$(pidof jmeter)
    if [ -z "\$PID" ]; then
        if test -f /tmp/start; then
            sudo aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE --duration-second 3600 > /tmp/irp-cred.txt
            export AWS_ACCESS_KEY_ID="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
            export AWS_SECRET_ACCESS_KEY="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
            export AWS_SESSION_TOKEN="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
            now=$(date +"%Y%m%d%I%M%p")
            jmeter -g /home/jmeter/results.jtl -o /home/jmeter/graphs
            aws s3 sync  /home/jmeter/graphs "s3://ugcupload-jmeter/$TENNANT/\$now/$HOSTNAME/graphs"
            aws s3api put-object --bucket ugcupload-jmeter --key "$TENNANT/\$now/$HOSTNAME/results.jtl" --body /home/jmeter/results.jtl
            echo "\$PID is empty"
        else
            echo "Report file not created"
        fi

        rm /tmp/start
    fi

fi
EOF
    echo "" >/tmp/start
    rm -rf /home/jmeter/graphs
    mkdir /home/jmeter/graphs
    rm -rf /home/jmeter/results.jtl
    touch /home/jmeter/results.jtl
    nohup bash -c "jmeter -n -GTENNANT=$TENNANT -t $1 -l /home/jmeter/results.jtl -Dserver.rmi.ssl.disable=true -R $3 >/home/jmeter/start.log 2>&1 &"
    echo "started test $1"
else
    echo "Test already running"
fi

exec "$@"