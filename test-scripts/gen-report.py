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

items_to_process = {}
jtl_items=[]
s3 = boto3.client('s3')

script = """
#!/bin/bash

echo "Combines all results from files called testresult*.jtl into one file called merged.jtl"
echo "If merged.jtl exists, it will be overridden"

cat /tmp/ugcupload/*.jtl > /tmp/ugcupload/merged.jtl

# Remove boundaries between tests
#sed 's_<\/testResults>__g' /tmp/ugcupload/merged.jtl > /tmp/ugcupload/sedmerged1
#sed 's_<?xml version=\"1.0\" encoding=\"UTF-8\"?>__g' /tmp/ugcupload/sedmerged1 > /tmp/ugcupload/sedmerged2
#sed 's_<testResults version=\"1.2\">__g' /tmp/ugcupload/sedmerged2 > /tmp/ugcupload/sedmerged3

# Add wrappers
#echo "</testResults>" >> /tmp/ugcupload/sedmerged3
#sed '1i <?xml version="1.0" encoding="UTF-8"?><testResults version="1.2">' /tmp/ugcupload/sedmerged3 > /tmp/ugcupload/merged.jtl
"""
def build_jmeter_graphs():
    if os.path.exists('/tmp/ugcupload/graph') and os.path.isdir('/tmp/ugcupload/graph'):
        shutil.rmtree('/tmp/ugcupload')
    os.mkdir('/tmp/ugcupload/graph')
    print("building graph")


    cmd = '${JMETER_HOME}/bin/jmeter -g /tmp/ugcupload/merged.jtl -o /tmp/ugcupload/graph'
    os.system(cmd)

def merge_jtl():
    os.mknod('/tmp/ugcupload/merge.sh')
    with open("/tmp/ugcupload/merge.sh", "w") as outfile:
        outfile.write(script)
    os.chmod("/tmp/ugcupload/merge.sh", 0o777)
    os.system("/tmp/ugcupload/merge.sh")
    
def download_objects():

    if os.path.exists('/tmp/ugcupload') and os.path.isdir('/tmp/ugcupload'):
        shutil.rmtree('/tmp/ugcupload')
    
    os.mkdir('/tmp/ugcupload')
    for i in jtl_items:
        s3.download_file('ugcupload-jmeter',i, "/tmp/ugcupload/"+str(uuid.uuid4())+".jtl")

def get_matching_s3_keys(bucket, prefix='', suffix=''):

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
                    print(key)
        try:
            kwargs['ContinuationToken'] = resp['NextContinuationToken']
        except KeyError:
            break

def get_items():
    print("get itsm")
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
    print(arguments['<items>'])
    process_arguments(arguments['<items>'])
    print(items_to_process)
    get_items()
    download_objects()
    merge_jtl()
    build_jmeter_graphs()