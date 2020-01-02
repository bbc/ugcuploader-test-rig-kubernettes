#!/usr/bin/env bash
working_dir="`pwd`"
pods="kubectl get pods -l jmeter_mode=master -n $1"
eval var=(\$\($pods\))
for i in "${var[@]}"
do
   if [[ $i == "jmeter-master"* ]]; then
     echo "hmm $i"
      echo "$working_dir/src/test" "/home/jmeter/"
      kubectl cp "$working_dir/src/test" "$i:/home/jmeter/" -n $1
   fi
done
