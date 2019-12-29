#!/bin/bash
set -e
cat > /home/jmeter/bin/check_if_ended.sh << EOF
#!/usr/bin/env bash

echo 1
if test -f /tmp/start; then
echo 2
    PID=\$(pidof jmeter)
    if [ -z "\$PID" ]; then
        sudo aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE --duration-second 1000 > /tmp/irp-cred.txt
        export AWS_ACCESS_KEY_ID="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
        export AWS_SECRET_ACCESS_KEY="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
        export AWS_SESSION_TOKEN="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
        now=$(date +"%Y%m%d%I%M%p")
        now="\$now/resutls.jtl"

        if test -f /tmp/start; then
             aws s3api put-object --bucket ugcupload-jmeter --key "$2/\$now" --body /home/jmeter/results.jtl
             echo "\$PID is empty"
        else
            echo "Report file not created"
        fi

        rm /tmp/start
    fi

fi
EOF
echo "" > /tmp/start
nohup jmeter -n -t $1 -l /home/jmeter/results.jtl -Dserver.rmi.ssl.disable=true -R `getent ahostsv4 jmeter-slaves-svc | cut -d' ' -f1 | sort -u | awk -v ORS=, '{print $1}' | sed 's/,$//'` &

