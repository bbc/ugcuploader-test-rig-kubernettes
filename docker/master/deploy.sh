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

rm -rf src
cp -R ../../src .
rm -rf jmeter-master
cp -R ../../jmeter-master .
REPO="$aws_acnt_num.dkr.ecr.$region.amazonaws.com/ugctestgrid/jmeter-master:latest"
echo "Repo: $REPO"
aws ecr delete-repository --force --repository-name ugctestgrid/jmeter-master
aws ecr create-repository --repository-name ugctestgrid/jmeter-master
docker build --pull --no-cache  -t ugctestgrid/jmeter-master .
docker tag ugctestgrid/jmeter-master:latest $REPO
docker push $REPO 
