#!/usr/bin/env bash
#Script writtent to stop a running jmeter master test
#Kindly ensure you have the necessary kubeconfig

export AWS_ACCESS_KEY_ID="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
export AWS_SECRET_ACCESS_KEY="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
export AWS_SESSION_TOKEN="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
export AWS_DEFAULT_REGION=eu-west-2

master_pod=`kubectl get po -n $1 | grep jmeter-master | awk '{print $1}'`
#
#
kubectl -n $1 exec -ti $master_pod bash /opt/apache-jmeter/bin/stoptest.sh
