#!/usr/bin/env bash

kubectl delete namespace ugcload-reporter
eksctl delete iamserviceaccount --name ugcupload-jmeter --namespace ugcload-reporter --cluster ugctestgrid
kubectl delete -f influxdb-pv.yaml -n ugcupload-reporter
kubectl delete -f influxdb-sc.yaml -n ugcupload-reporter
kubectl delete -f chronograf-pv.yaml -n ugcupload-reporter
kubectl delete -f chronograf-sc.yaml -n ugcupload-reporter
kubectl delete -f grafana-pv.yaml -n ugcupload-reporter
kubectl delete -f grafana-sc.yaml -n ugcupload-reporter

