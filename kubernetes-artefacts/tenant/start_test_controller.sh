#!/usr/bin/env bash

if [[ ! -z "$AWS_ROLE_ARN" ]]; 
then
   uuid=$(python -c 'import sys,uuid; sys.stdout.write(uuid.uuid4().hex)')

   aws sts assume-role-with-web-identity --role-arn $AWS_ROLE_ARN --role-session-name mh9test --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE --duration-second 3600  > "/tmp/$uuid.txt"
   aak='cat /tmp/$uuid.txt | jq -r ".Credentials.AccessKeyId"'
   sak='cat /tmp/$uuid.txt | jq -r ".Credentials.SecretAccessKey"'
   st='cat /tmp/$uuid.txt | jq -r ".Credentials.SessionToken"'
   export AWS_ACCESS_KEY_ID=$(eval "$aak")
   export AWS_SECRET_ACCESS_KEY=$(eval "$sak")
   export AWS_SESSION_TOKEN=$(eval "$st")
   export AWS_DEFAULT_REGION=eu-west-2
   rm "/tmp/$uuid.txt"
fi


k_p="kubectl get pods -n $2"
function check_if_all_started {

echo "k_p:$k_p"
eval kp=(\$\($k_p\))
ignore=5
count=0
inner=0
innerin=0
mod=0
found=0
for i in "${kp[@]}"
do
   if (($count >= $ignore))
   then
    let "mod=inner%5"
    if (($mod == 0))
    then
       let "innerin=0"
    fi

    if (($innerin == 2))
    then

        if [ "$i" != "Running" ]; then
            echo "$innerin $i"
            let  "found=found+1"
        fi
    fi   
    let "innerin=innerin+1"
    let "inner=inner+1"
   fi
   let "count=count+1"
done 
    
}

kubectl scale --replicas=$4 deployment/jmeter-slave -n $2
check_if_all_started
until ((  $found < 1  ))
do
  sleep 5
  check_if_all_started
  if (($found < 1))
  then
     break
  fi
done
echo "done"

#working_dir="/home/control"
working_dir="/Users/baahk01/workspace/ugcuploader-test-kubernettes"

echo "ork = $working_dir"

jmx="$working_dir/src/test/$1"
[ -n "$jmx" ] || read -p 'Enter path to the jmx file ' jmx

if [ ! -f "$jmx" ];
then
    echo "Test script file was not found: $jmx"
    echo "Kindly check and input the correct file path"
    exit
fi

# Copy bandwidth config to slaves
slave_pods="kubectl get pods -l jmeter_mode=slave -n $2"
eval slave_var=(\$\($slave_pods\))
for i in "${slave_var[@]}"
do
   if [[ $i == "jmeter-slave"* ]]; then
     echo "hmm $i"
      kubectl cp "$working_dir/config/bandwidth/$3/bandwidth.csv" "$i:/opt/apache-jmeter/bin/jmeter.properties" -n $2 &
      sleep 1
      killall kubectl
      kubectl cp "$working_dir/data" "$i:/" -n $2 &
      sleep 1
      killall kubectl
      kubectl exec -it -n $2 $i -- bash -c /start.sh &
      sleep 2
      killall kubectl
   fi
done

test_to_run="$1"

master_pod=`kubectl get po -n $2 | grep jmeter-master | awk '{print $1}'`

# Copy test to master
path=${test_to_run%/*} 
root=$(echo "$path" | cut -d "/" -f1)
kubectl exec -it -n $2 $master_pod  -- bash -c "rm -rf test/$root"
kubectl exec -it -n $2 $master_pod  -- bash -c "mkdir test/$path" 
kubectl cp "$working_dir/src/test/$test_to_run" "$master_pod:/home/jmeter/test/$path" -n $2
echo "Starting Jmeter load test $test_to_run for $2 running on $master_pod  "

kubectl exec -it -n $2 $master_pod -- bash -c "/home/jmeter/bin/load_test.sh /home/jmeter/test/$test_to_run $2" 
