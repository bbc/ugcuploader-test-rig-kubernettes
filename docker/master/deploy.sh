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

rm -rf src
cp -R ../../src .
REPO="$aws_acnt_num.dkr.ecr.$region.amazonaws.com/ugcloadtest/jmeter-master:latest"
echo "Repo: $REPO"
aws ecr delete-repository --force --repository-name ugcloadtest/jmeter-master
aws ecr create-repository --repository-name ugcloadtest/jmeter-master
docker build --pull --no-cache  -t ugcloadtest/jmeter-master .
docker tag ugcloadtest/jmeter-master:latest $REPO
docker push $REPO 
