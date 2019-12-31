#!/usr/bin/env bash
#Script created to launch Jmeter tests directly from the current terminal without accessing the jmeter master pod.
#It requires that you supply the path to the jmx file
#After execution, test script jmx file may be deleted from the pod itself but not locally.

working_dir="`pwd`"

echo "ork = $working_dir"

jmx="$working_dir/src/test/$1"
[ -n "$jmx" ] || read -p 'Enter path to the jmx file ' jmx

if [ ! -f "$jmx" ];
then
    echo "Test script file was not found: $jmx"
    echo "Kindly check and input the correct file path"
    exit
fi

test_name="$1"

#Get Master pod details

master_pod=`kubectl get po -n $2 | grep jmeter-master | awk '{print $1}'`

echo "Starting Jmeter load test $test_name for $2 running on $master_pod  "

kubectl exec -ti -n $2 $master_pod -- /bin/bash /home/jmeter/bin/load_test.sh "/home/jmeter/test/$test_name" $2 