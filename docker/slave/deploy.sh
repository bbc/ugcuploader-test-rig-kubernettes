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

rm -rf fileupload
cp -R ../../fileupload .

REPO="$aws_acnt_num.dkr.ecr.$region.amazonaws.com/ugctestgrid/jmeter-slave:latest"
echo "Repo: $REPO"
aws ecr delete-repository --force --repository-name ugctestgrid/jmeter-slave
aws ecr create-repository --repository-name ugctestgrid/jmeter-slave
docker build --pull --no-cache  -t ugctestgrid/jmeter-slave .
docker tag ugctestgrid/jmeter-slave:latest $REPO
docker push $REPO 