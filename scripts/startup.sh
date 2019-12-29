#!/bin/bash

if [ ! -d "venv" ] 
then
    rm -rf venv
    python3 --version
    python3 -m venv venv
    source ./venv/bin/activate 
    pip install --upgrade pip 
    pip install -r requirements.txt 
fi
#nohup ./venv/bin/python cleanup/database_tunell_jmeter.py int ugc-cleaner 5433 &


HEAP="-Xms1g -Xmx1g -XX:MaxMetaspaceSize=512m"
export _JAVA_OPTIONS="-Xms512m -Xmx1024m"
JAVA_OPTIONS="-Djavax.net.ssl.keyStoreType=pkcs12 -Djavax.net.ssl.keyStore=/Users/baahk01/workspace/bbc_cert.p12 -Djavax.net.ssl.keyStorePassword=vagrant"

$JMETER_HOME/bin/jmeter.sh -n -N 127.0.0.1 $JAVA_OPTIONS -t upload/upload.jmx

