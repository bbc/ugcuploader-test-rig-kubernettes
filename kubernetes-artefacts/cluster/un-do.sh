#!/usr/bin/env bash

eksctl delete cluster -f cluster.yaml
POLICY_ARN="arn:aws:iam::$1:policy/ugcupload-eks-jmeter-policy"
echo $POLICY_ARN
aws iam delete-policy --policy-arn $POLICY_ARN
