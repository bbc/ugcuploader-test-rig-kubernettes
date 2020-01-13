#!/usr/bin/env python3

"""Generate Report

Usage:
   gen_report.py <items>

Arguments:
    items       a comma separated list of tennet and date eg national-moments=201912310149PM,children=202001011136PM

"""
from docopt import docopt
import boto3
import uuid
import tempfile
import os
import shutil
import glob
import subprocess
import sys
from string import Template
from datetime import date


items_to_process = {}
jtl_items = []
sts = boto3.client('sts')

fileName = str(uuid.uuid4())

script = """
#!/bin/bash

echo "Combines all results from files called testresult*.jtl into one file called merged.jtl"
echo "If merged.jtl exists, it will be overridden"

cat /home/control/graphs/$filename/*.jtl > /home/control/graphs/$filename/merged.jtl

# Remove boundaries between tests
# sed 's_<\/testResults>__g' /tmp/ugcupload/merged.jtl > /tmp/ugcupload/sedmerged1
# sed 's_<?xml version=\"1.0\" encoding=\"UTF-8\"?>__g' /tmp/ugcupload/sedmerged1 > /tmp/ugcupload/sedmerged2
# sed 's_<testResults version=\"1.2\">__g' /tmp/ugcupload/sedmerged2 > /tmp/ugcupload/sedmerged3

# Add wrappers
# echo "</testResults>" >> /tmp/ugcupload/sedmerged3
# sed '1i <?xml version="1.0" encoding="UTF-8"?><testResults version="1.2">' /tmp/ugcupload/sedmerged3 > /tmp/ugcupload/merged.jtl
"""


def build_jmeter_graphs():
    jm = os.environ['JMETER_HOME']
    today = date.today()
    d1 = today.strftime("%Y-%m-%d-%s")
    u = str(uuid.uuid4())
    cmd = 'sudo {1}/bin/jmeter -g /home/control/graphs/{0}/merged.jtl -o /var/www/localhost/htdocs/{2}-{3}'.format(fileName, jm, d1, u)
    os.system(cmd)


def merge_jtl():
    os.mknod('/home/control/graphs/{0}/merge.sh'.format(fileName))
    with open("/home/control/graphs/{0}/merge.sh".format(fileName), "w") as outfile:
        s = Template(script)
        o = s.substitute(filename=fileName)
        outfile.write(o)
    os.chmod("/home/control/graphs/{0}/merge.sh".format(fileName), 0o777)
    os.system("/home/control/graphs/{0}/merge.sh".format(fileName))


def download_objects():

    webTokenFile = os.environ['AWS_WEB_IDENTITY_TOKEN_FILE']
    if webTokenFile:
       with open(webTokenFile, 'r') as file:
          webToken = file.read().replace('\n', '')
       roleArn = os.environ["AWS_ROLE_ARN"]
       response = sts.assume_role_with_web_identity(
             RoleArn = roleArn,
             RoleSessionName=str(uuid.uuid4()),
             WebIdentityToken=webToken,
             DurationSeconds=3600)

       accessKeyId = response['Credentials']['AccessKeyId']
       secretAccessKey = response['Credentials']['SecretAccessKey']
       sessionToken = response['Credentials']['SessionToken']
    
       s3 = boto3.client(
        's3',
        aws_access_key_id=accessKeyId,
        aws_secret_access_key=secretAccessKey,
        aws_session_token=sessionToken,
       )

    os.makedirs("/home/control/graphs/{0}".format(fileName), exist_ok=True)
    for i in jtl_items:
        fn="/home/control/graphs/{0}/{1}.jtl".format(fileName,str(uuid.uuid4()))
        try:
            f=open(fn, 'w+b')
            response = s3.get_object(Bucket='ugcupload-jmeter',Key=i)
            contents = response['Body'].read()
            f.write(contents)
            f.flush()
            f.close()
        except:
           print("Unexpected error:{0}".format(sys.exc_info()[0]))


def get_matching_s3_keys(bucket, prefix='', suffix=''):

    webTokenFile = os.environ['AWS_WEB_IDENTITY_TOKEN_FILE']
    if webTokenFile:
       with open(webTokenFile, 'r') as file:
          webToken = file.read().replace('\n', '')
       roleArn = os.environ["AWS_ROLE_ARN"]
       response = sts.assume_role_with_web_identity(
             RoleArn = roleArn,
             RoleSessionName=str(uuid.uuid4()),
             WebIdentityToken=webToken,
             DurationSeconds=3600)

       accessKeyId = response['Credentials']['AccessKeyId']
       secretAccessKey = response['Credentials']['SecretAccessKey']
       sessionToken = response['Credentials']['SessionToken']
    
       s3 = boto3.client(
        's3',
        aws_access_key_id=accessKeyId,
        aws_secret_access_key=secretAccessKey,
        aws_session_token=sessionToken,
       )
       kwargs = {'Bucket': bucket}

       if isinstance(prefix, str):
            kwargs['Prefix'] = prefix

       while True:
            resp = s3.list_objects_v2(**kwargs)
            count = resp['KeyCount']
            if count > 0:
                for obj in resp['Contents']:
                    key = obj['Key']
                    if key.endswith(suffix):
                        jtl_items.append(key)
            try:
                kwargs['ContinuationToken'] = resp['NextContinuationToken']
            except KeyError:
                break

def get_items():
    for k, v in items_to_process.items():
        for i in v:
            get_matching_s3_keys('ugcupload-jmeter',k+"/"+i,"jtl")
        

def process_arguments(items):
    l = items.split(",")
    for item in l:
        i = item.split("=")
        try:
            items_to_process[i[0]].append(i[1])
        except:
            items_to_process[i[0]]=[i[1]]
    

if __name__ == '__main__':
    arguments = docopt(__doc__)
    process_arguments(arguments['<items>'])
    get_items()
    download_objects()
    merge_jtl()
    build_jmeter_graphs()
