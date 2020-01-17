#!/usr/bin/env bash

kubectl delete namespace ugcload-reporter
eksctl delete iamserviceaccount --name ugcupload-jmeter --namespace ugcload-reporter --cluster ugctestgrid
