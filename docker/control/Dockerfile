# Base build image
FROM python:3.7-alpine

USER 0
# Install some dependencies needed to build the project
RUN apk update && apk --no-cache add curl git jq bash openssh openrc rsyslog rsync gcc g++ libc-dev bash go bzr git mercurial subversion openssh-client ca-certificates lighttpd redis

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
	&& update-ca-certificates \
	&& apk add --update openjdk8-jre tzdata \
	&& apk add --no-cache nss \
	&& rm -rf /var/cache/apk/* \
	&& mkdir -p /tmp/dependencies  \
	&& curl -L --silent ${JMETER_DOWNLOAD_URL} >  /tmp/dependencies/apache-jmeter-${JMETER_VERSION}.tgz  \
	&& mkdir -p /opt  \
	&& tar -xzf /tmp/dependencies/apache-jmeter-${JMETER_VERSION}.tgz -C /opt  \
	&& rm -rf /tmp/dependencies

# Set global PATH such that "jmeter" command is found
ENV PATH $PATH:$JMETER_BIN
COPY plugins/* /opt/apache-jmeter-$JMETER_VERSION/lib/ext/
COPY conf/* /opt/apache-jmeter-$JMETER_VERSION/bin/
COPY lib/* /opt/apache-jmeter-$JMETER_VERSION/lib/
COPY client.jks /

RUN apk add --no-cache shadow sudo && \
    if [ -z "`getent group 1000`" ]; then \
      addgroup -S -g 1000 control; \
    else \
      groupmod -n control `getent group 1000 | cut -d: -f1`; \
    fi && \
    if [ -z "`getent passwd 1000`" ]; then \
      adduser -S -u 1000 -G control -s /bin/sh control; \
    else \
      usermod -l control -g control -d /home/control -m `getent passwd control | cut -d: -f1`; \
    fi && \
    echo "control ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/control && \
    chmod 0440 /etc/sudoers.d/control



WORKDIR /home/control
USER 1000
COPY etc /etc
RUN mkdir /home/control/admin

COPY admin admin
USER 0
RUN chown -R control:control admin
USER 1000
# add credentials on build
#ARG SSH_PRIVATE_KEY
#RUN mkdir ~/.ssh/
#RUN echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa
WORKDIR /home/control/admin
# Force the go compiler to use modules 
ENV GO111MODULE=on

#This is the ‘magic’ step that will download all the dependencies that are specified in 
# the go.mod and go.sum file.

# Because of how the layer caching system works in Docker, the go mod download 
# command will _ only_ be re-run when the go.mod or go.sum file change 
# (or when we add another docker instruction this line) 
RUN go mod download

# This image builds the weavaite server
# Here we copy the rest of the source code

# And compile the project


RUN go build -o ./bin/admin  ./cmd/ugcupload/main.go
WORKDIR /home/control
USER 0
RUN curl --silent --location "https://github.com/weaveworks/eksctl/releases/download/latest_release/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
RUN mv /tmp/eksctl /usr/local/bin
RUN rc-update add sshd
RUN rc-update add rsyslog boot
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.1/bin/linux/amd64/kubectl
RUN chmod u+x kubectl && mv kubectl /bin/kubectl

RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip"
RUN unzip awscli-bundle.zip
RUN ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws
RUN rm ./awscli-bundle.zip
RUN rm -rf ./awscli-bundle/

RUN curl -o aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/1.14.6/2019-08-22/bin/linux/amd64/aws-iam-authenticator
RUN chmod +x ./aws-iam-authenticator

COPY crontab.txt crontab.txt
RUN chmod 0644 /home/control/crontab.txt
RUN crontab -u control /home/control/crontab.txt

USER 1000
COPY bin/ /usr/local/bin
COPY test-scripts/ /usr/local/bin
RUN mkdir tenant
COPY tenant tenant
RUN mkdir src
COPY test/ src/test
RUN mkdir config
COPY config config
RUN mkdir data
COPY entry.sh /usr/local/bin
COPY lighttpd.conf /etc/lighttpd/lighttpd.conf
RUN mkdir graphs

USER 0
RUN chmod 0777 /usr/local/bin/gen-env.py
RUN chown -R control:control tenant
RUN chown -R control:control src
RUN chown -R control:control config
RUN chown -R control:control data
RUN chmod 0777 /usr/local/bin/entry.sh
RUN pip install docopt
RUN pip install boto3
RUN mv ./aws-iam-authenticator /usr/local/bin

EXPOSE 1323
ENTRYPOINT ["/usr/local/bin/entry.sh"]
