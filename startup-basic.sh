#!/bin/bash


HEAP="-Xms1g -Xmx1g -XX:MaxMetaspaceSize=512m"
export _JAVA_OPTIONS="-Xms512m -Xmx1024m"
JAVA_OPTIONS="-Djavax.net.ssl.keyStoreType=pkcs12 -Djavax.net.ssl.keyStore=/Users/baahk01/workspace/bbc_cert.p12 -Djavax.net.ssl.keyStorePassword=vagrant"

$JMETER_HOME/bin/jmeter.sh -N 127.0.0.1 $JAVA_OPTIONS -t src/test/upload/upload.jmx

