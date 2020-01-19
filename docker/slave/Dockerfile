# Use vinsdocker base image
FROM 546933502184.dkr.ecr.eu-west-2.amazonaws.com/ugctestgrid/jmeter-base:latest
MAINTAINER Kodjo Baah<kafriyie01@bbc.co.uk>

# Ports to be exposed from the container for JMeter Slaves/Server
EXPOSE 1099 50000

RUN apk add go
RUN mkdir /data
COPY data/ /data
RUN mkdir /config
COPY config/ /config
COPY entrypoint.sh /
COPY client.jks /
COPY start.sh /

RUN mkdir /fileupload
COPY fileupload fileupload

WORKDIR /fileupload


# Force the go compiler to use modules 
ENV GO111MODULE=on


# Because of how the layer caching system works in Docker, the go mod download 
# command will _ only_ be re-run when the go.mod or go.sum file change 
# (or when we add another docker instruction this line) 
RUN go mod download

# This image builds the weavaite server
# Here we copy the rest of the source code
# And compile the project

RUN go build -o upload cmd/ugcupload/main.go

ENTRYPOINT ["/entrypoint.sh"]
