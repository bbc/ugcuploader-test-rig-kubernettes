#!/usr/bin/env bash

declare RESULT=($(eksctl utils describe-stacks --cluster ugctestgrid | grep StackId))  
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

POLICY_ARN="arn:aws:iam::$aws_acnt_num:policy/ugcupload-eks-control-policy "
echo $POLICY_ARN
kubectl create namespace control
eksctl create iamserviceaccount --name ugcupload-control --namespace control  --cluster ugctestgrid --attach-policy-arn $POLICY_ARN --approve --override-existing-serviceaccounts
kubectl create -f clusterolebinding.yaml
kubectl create -n control -f control.yaml
