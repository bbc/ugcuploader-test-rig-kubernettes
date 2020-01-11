#!/usr/bin/env bash

s_p="kubectl get rs -n $1 | grep jmeter-slave | awk '{print \$1}'"
eval sp=\$\($s_p\)
echo slave-pod=$sp

s_r="kubectl scale --replicas=$2 $sp -n $1"
eval sr=\$\($s_r\)
echo slave-rep=$sr
#slave_pod=`kubectl get rs -n jame | grep jmeter-slave | awk '{print $1}'`
#kubectl scale --replicas=$2 $slave_pod