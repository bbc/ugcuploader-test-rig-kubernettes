#!/usr/bin/env bash
working_dir="`pwd`"
pods="kubectl get pods -l jmeter_mode=slave -n $1"
eval var=(\$\($pods\))
for i in "${var[@]}"
do
   if [[ $i == "jmeter-slaves"* ]]; then
     echo "hmm $i"
      kubectl cp "$working_dir/docker/base/conf/user.properties" "$1/$i:/opt/apache-jmeter/bin"
   fi
done
