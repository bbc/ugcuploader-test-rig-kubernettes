# Use vinsdocker base image
FROM 546933502184.dkr.ecr.eu-west-2.amazonaws.com/ugcloadtest/jmeter-base:latest
MAINTAINER Kodjo Baah<kafriyie01@bbc.co.uk>

# Ports to be exposed from the container for JMeter Slaves/Server
EXPOSE 1099 50000

RUN mkdir /data
COPY data/ /data
RUN mkdir /config
COPY config/ /config
COPY entrypoint.sh /
COPY client.jks /


# Application to run on starting the container
#ENTRYPOINT $JMETER_HOME/bin/jmeter-server \
#                        -Dserver.rmi.localport=50000 \
#                        -Dserver_port=1099
