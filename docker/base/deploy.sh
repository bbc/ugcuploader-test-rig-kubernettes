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

REPO="$aws_acnt_num.dkr.ecr.$region.amazonaws.com/ugcloadtest/jmeter-base:latest"
aws ecr delete-repository --force --repository-name ugcloadtest/jmeter-base
aws ecr create-repository --repository-name ugcloadtest/jmeter-base 
docker build --no-cache -t ugcloadtest/jmeter-base .
docker tag ugcloadtest/jmeter-base:latest $REPO
docker push $REPO
