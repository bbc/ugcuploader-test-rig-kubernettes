# inspired https://github.com/justb4/docker-jmeter/
FROM python:3.7-alpine

MAINTAINER Kodjo Baah<kafriyie01@bbc.co.uk>

ARG JMETER_VERSION="5.1.1"
ENV JMETER_HOME /opt/apache-jmeter-${JMETER_VERSION}
ENV	JMETER_BIN	${JMETER_HOME}/bin
ENV	JMETER_DOWNLOAD_URL  https://archive.apache.org/dist/jmeter/binaries/apache-jmeter-${JMETER_VERSION}.tgz
RUN ln -s /opt/apache-jmeter-${JMETER_VERSION} /opt/apache-jmeter
# Install extra packages
# See https://github.com/gliderlabs/docker-alpine/issues/136#issuecomment-272703023
# Change TimeZone TODO: TZ still is not set!
ARG TZ="Europe/Amsterdam"
RUN    apk update \
	&& apk upgrade \
	&& apk add ca-certificates \
	&& update-ca-certificates \
	&& apk add --update openjdk8-jre tzdata curl unzip bash \
	&& apk add --no-cache nss \
	&& rm -rf /var/cache/apk/* \
	&& mkdir -p /tmp/dependencies  \
	&& curl -L --silent ${JMETER_DOWNLOAD_URL} >  /tmp/dependencies/apache-jmeter-${JMETER_VERSION}.tgz  \
	&& mkdir -p /opt  \
	&& tar -xzf /tmp/dependencies/apache-jmeter-${JMETER_VERSION}.tgz -C /opt  \
	&& rm -rf /tmp/dependencies

RUN apk add rsync openssh
# Set global PATH such that "jmeter" command is found
ENV PATH $PATH:$JMETER_BIN
COPY plugins /opt/apache-jmeter-$JMETER_VERSION/lib/ext
COPY conf /opt/apache-jmeter-$JMETER_VERSION/bin
COPY lib /opt/apache-jmeter-$JMETER_VERSION/lib
copy client.jks /
RUN mkdir /test-output
