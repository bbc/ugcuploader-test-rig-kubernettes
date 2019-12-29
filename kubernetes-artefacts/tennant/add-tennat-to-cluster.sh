#!/usr/bin/env bash

POLICY_ARN="arn:aws:iam::$2:policy/ugcupload-eks-jmeter-policy"
echo $POLICY_ARN
kubectl create namespace $1 
eksctl create iamserviceaccount --name ugcupload-jmeter --namespace $1  --cluster ugcloadtest --attach-policy-arn $POLICY_ARN --approve --override-existing-serviceaccounts
cat jmeter-slaves.yaml.template | awk -v "act_num=$2" '{gsub(/AWS_ACCOUNT_NUMBER/,act_num)}1' | awk -v "region=$3" '{gsub(/AWS_REGION/,region)}1' > jmeter-slaves.yaml
cat jmeter-master.yaml.template | awk -v "act_num=$2" '{gsub(/AWS_ACCOUNT_NUMBER/,act_num)}1' | awk -v "region=$3" '{gsub(/AWS_REGION/,region)}1' > jmeter-master.yaml

kubectl create -n $1 -f jmeter-master.yaml
kubectl create -n $1 -f jmeter-slaves.yaml