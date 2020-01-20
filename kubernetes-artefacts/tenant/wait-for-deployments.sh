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

found=0
k_p="kubectl -n $1 get pods | awk '{print \$3}'"
function check_if_all_started {
IFS=$'\n'
echo "k_p:$k_p"
eval kp=(\$\($k_p\))
for i in "${kp[@]}"
do
        if [ "$i" == "Running" ]; then
            echo "Found $i"
            let "found=found+1"
        fi
done 
    
}

while [ $found -lt $2 ]
do
    check_if_all_started
    if [$found -lt $2 ];then 
        lef "found=0"
    if
done

