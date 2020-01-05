#!/usr/bin/env bash

kubectl delete namespace control
eksctl delete iamserviceaccount --name  ugcupload-control --namespace control --cluster ugcloadtest
