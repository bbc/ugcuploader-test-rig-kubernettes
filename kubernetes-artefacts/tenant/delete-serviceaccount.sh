#!/usr/bin/env bash

uuid=$(python -c 'import sys,uuid; sys.stdout.write(uuid.uuid4().hex)')

aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB
aak='cat /tmp/$uuid.txt | jq -r ".Credentials.AccessKeyId"'
sak='cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey"'
st='cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken"'
export AWS_ACCESS_KEY_ID=$(eval "$aak")
export AWS_SECRET_ACCESS_KEY=$(eval "$sak")
export AWS_SESSION_TOKEN=$(eval "$st")
export AWS_DEFAULT_REGION=eu-west-2
rm "/tmp/$uuid.txt"
eksctl delete iamserviceaccount --name ugcupload-jmeter --namespace $1 --cluster ugcloadtest
kubectl delete namespace $1 
