#!/usr/bin/env bash
declare RESULT=($(eksctl utils describe-stacks --cluster ugcloadtest | grep StackId))  
for i in "${RESULT[@]}"
do
    var="${i%\"}"
    var="${var#\"}"
    if [[ $var == "arn:aws:cloudformation"* ]]; then
        arrIN=(${var//:/ })
        region=${arrIN[3]}
        aws_acnt_num=${arrIN[4]}
    fi
   # do whatever on $i
done

echo $region
echo $aws_acnt_num

POLICY_ARN="arn:aws:iam::$aws_acnt_num:policy/ugcupload-eks-jmeter-policy"
echo $POLICY_ARN
kubectl create namespace ugcload-reporter
eksctl create iamserviceaccount --name ugcupload-jmeter --namespace ugcload-reporter --cluster ugcloadtest --attach-policy-arn $POLICY_ARN --approve --override-existing-serviceaccounts
kubectl create -n ugcload-reporter -f ./grafana.yaml
kubectl create -n ugcload-reporter -f ./influxdb.yaml