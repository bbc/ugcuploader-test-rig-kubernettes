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

rm -rf test-scripts
cp -R ../../test-scripts .
rm -rf tennant
cp -R ../../kubernetes-artefacts/tennant .
rm -rf test 
cp -R ../../src/test .
rm -rf config
cp -R ../../config .
rm -rf data
cp -R ../../data .
rm -rf admin
cp -R ../../admin .


REPO="$aws_acnt_num.dkr.ecr.$region.amazonaws.com/ugcloadtest/control:latest"
aws ecr delete-repository --force --repository-name ugcloadtest/control
aws ecr create-repository --repository-name ugcloadtest/control
sudo docker build --no-cache -t ugcloadtest/control .
sudo docker tag ugcloadtest/control:latest $REPO
sudo docker push $REPO
