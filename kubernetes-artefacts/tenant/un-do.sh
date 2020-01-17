#!/usr/bin/env bash

eksctl delete iamserviceaccount --name ugcupload-jmeter --namespace $1 --cluster ugctestgrid
kubectl delete namespace $1 
