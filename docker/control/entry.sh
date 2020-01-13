#!/bin/bash
set -e

echo "export AWS_WEB_IDENTITY_TOKEN_FILE=$AWS_WEB_IDENTITY_TOKEN_FILE" | sudo tee -a /etc/profile
echo "export AWS_ROLE_ARN=$AWS_ROLE_ARN" | sudo tee -a /etc/profile
echo "export AWS_DEFAULT_REGION=eu-west-2" | sudo tee -a /etc/profile

sudo rc-status
sudo touch /run/openrc/softlevel
sudo service rsyslog start
sudo service sshd start
echo "PasswordAuthentication yes" | sudo tee -a /etc/ssh/sshd_config
echo "ClientAliveInterval 60" | sudo tee -a /etc/ssh/sshd_config
echo "sshd : ALL" | sudo tee -a  /etc/hosts.allow

PID=$(ps | grep "kubectl port-forward" | grep -v grep| awk {'print$1'})
echo $PID
if [ ! -z "$PID" ]; then
    kill -9 $PID
fi
sudo aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE --duration-second 3600 > /tmp/irp-cred.txt
export AWS_ACCESS_KEY_ID="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
export AWS_SECRET_ACCESS_KEY="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
export AWS_SESSION_TOKEN="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
export AWS_DEFAULT_REGION=eu-west-2

aws eks --region eu-west-2 update-kubeconfig --name ugcloadtest
nohup kubectl port-forward --address 0.0.0.0 -n weave "$(kubectl get -n weave pod --selector=weave-scope-component=app -o jsonpath='{.items..metadata.name}')" 4040 &> weavscope.out&
nohup go run /home/control/admin/cmd/ugcupload/main.go &> admincontroller.out&

sudo cat >/home/control/start_weavescope.sh<<EOF
#!/usr/bin/env bash

PID=$(ps | grep "kubectl port-forward" | grep -v grep| awk {'print$1'})
echo $PID
if [ ! -z "\$PID" ]; then
    kill -9 $PID
fi
export AWS_ACCESS_KEY_ID="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
export AWS_SECRET_ACCESS_KEY="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
export AWS_SESSION_TOKEN="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
export AWS_DEFAULT_REGION=eu-west-2

nohup kubectl port-forward --address 0.0.0.0 -n weave "$(kubectl get -n weave pod --selector=weave-scope-component=app -o jsonpath='{.items..metadata.name}')" 4040 &> weavscope.out&

EOF
sudo chmod 0777 /home/control/start_weavescope.sh
sudo mv /home/control/start_weavescope.sh /usr/local/bin

sudo cat >/home/control/refresh_creds.sh<<EOF
#!/usr/bin/env bash
sudo aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE --duration-second 3600 > /tmp/irp-cred.txt
EOF

sudo chmod 0777 /home/control/refresh_creds.sh
sudo mv /home/control/refresh_creds.sh /usr/local/bin

sudo cat >/home/control/start_admin_controller.sh<<EOF
#!/usr/bin/env bash

PID=$(pidof go)
echo $PID
if [ ! -z "\$PID" ]; then
    kill -9 $PID
fi
export AWS_ACCESS_KEY_ID="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
export AWS_SECRET_ACCESS_KEY="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
export AWS_SESSION_TOKEN="\$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
export AWS_DEFAULT_REGION=eu-west-2

nohup go run /home/control/admin/cmd/ugcupload/main.go &> admincontroller.out&
EOF

sudo chmod 0777 /home/control/start_admin_controller.sh
sudo mv /home/control/start_admin_controller.sh /usr/local/bin

sudo service sshd start
sudo crond  -d 8 
sudo rc-service lighttpd start
echo "tart $@"
# Hand off to the CMD
exec "$@"