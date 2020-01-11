#!/usr/bin/env bash

eksctl create iamserviceaccount --name ugcupload-jmeter --namespace $1  --cluster ugcloadtest --attach-policy-arn $2 --approve --override-existing-serviceaccounts
