#!/usr/bin/env bash

POLICY_ARN="arn:aws:iam::$1:policy/ugcupload-eks-jmeter-policy"
echo $POLICY_ARN
kubectl create namespace ugcload-reporter
eksctl create iamserviceaccount --name ugcupload-jmeter --namespace ugcload-reporter --cluster ugcloadtest --attach-policy-arn $POLICY_ARN --approve --override-existing-serviceaccounts
kubectl create -n ugcload-reporter -f ./grafana.yaml
kubectl create -n ugcload-reporter -f ./influxdb.yaml