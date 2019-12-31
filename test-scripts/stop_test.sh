#!/usr/bin/env bash
#Script writtent to stop a running jmeter master test
#Kindly ensure you have the necessary kubeconfig

master_pod=`kubectl get po -n $1 | grep jmeter-master | awk '{print $1}'`
#
#
kubectl -n $1 exec -ti $master_pod bash /opt/apache-jmeter/bin/stoptest.sh
