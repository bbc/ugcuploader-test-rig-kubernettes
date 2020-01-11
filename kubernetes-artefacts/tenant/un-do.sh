#!/usr/bin/env bash

eksctl delete iamserviceaccount --name ugcupload-jmeter --namespace $1 --cluster ugcloadtest
kubectl delete namespace $1 
