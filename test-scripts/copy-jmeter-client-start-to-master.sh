#!/usr/bin/env bash
working_dir="`pwd`"
pods="kubectl get pods -l jmeter_mode=master -n $1"
eval var=(\$\($pods\))
for i in "${var[@]}"
do
   if [[ $i == "jmeter-master"* ]]; then
      echo "hmm $i"
      echo "$working_dir/docker/master/bin/load_test.sh" "/home/jmeter/bin"
      kubectl cp "$working_dir/docker/master/bin/load_test.sh" "$i:/home/jmeter/bin" -n $1
   fi
done