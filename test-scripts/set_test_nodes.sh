#!/usr/bin/env bash



function check_if_all_started {
# Copy bandwidth config to slaves
k_p="kubectl get pods -n jamie"
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

kubectl scale deployment.v1.apps/jmeter-slaves --replicas=$2 -n $1
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