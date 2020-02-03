# Use vinsdocker base image
FROM 546933502184.dkr.ecr.eu-west-2.amazonaws.com/ugctestgrid/jmeter-base:latest
MAINTAINER Kodjo Baah<kafriyie01@bbc.co.uk>

# Ports to be exposed from the container for JMeter Slaves/Server
EXPOSE 1099 50000 8778

RUN apk add go openjdk8
RUN mkdir /data
COPY data/ /data
RUN mkdir /config
COPY config/ /config
COPY entrypoint.sh /
COPY client.jks /
COPY start.sh /

RUN mkdir /fileupload
COPY fileupload fileupload

COPY jolokia-jvm-1.6.2-agent.jar /fileupload
WORKDIR /fileupload
ENV GO111MODULE=on
RUN go mod download
RUN go build -o upload cmd/ugcupload/main.go

ENTRYPOINT ["/entrypoint.sh"]
