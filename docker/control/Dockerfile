# Base build image
FROM python:3.7-alpine

USER root
# Install some dependencies needed to build the project
RUN apk update && apk --no-cache add curl git jq bash openssh openrc rsyslog rsync gcc g++ libc-dev bash go bzr git mercurial subversion openssh-client ca-certificates

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
RUN mkdir /home/control/admin


# add credentials on build
ARG SSH_PRIVATE_KEY
RUN mkdir ~/.ssh/
RUN echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa

# Force the go compiler to use modules 
ENV GO111MODULE=on

# We want to populate the module cache based on the go.{mod,sum} files. 
COPY admin/go.mod .
COPY admin/go.sum .

#This is the ‘magic’ step that will download all the dependencies that are specified in 
# the go.mod and go.sum file.

# Because of how the layer caching system works in Docker, the go mod download 
# command will _ only_ be re-run when the go.mod or go.sum file change 
# (or when we add another docker instruction this line) 
RUN go mod download

# This image builds the weavaite server
# Here we copy the rest of the source code
COPY admin admin
# And compile the project

#RUN CGO_ENABLED=1 GOOS=linux go build -o bin/admin -a -tags netgo -ldflags '-w -extldflags "-static"' ./admin/cmd/ugcupload/main.go

COPY etc /etc

RUN curl --silent --location "https://github.com/weaveworks/eksctl/releases/download/latest_release/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
RUN sudo mv /tmp/eksctl /usr/local/bin
RUN sudo rc-update add sshd
#RUN sudo touch /run/openrc/softlevel
RUN sudo rc-update add rsyslog boot
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.1/bin/linux/amd64/kubectl
RUN sudo chmod u+x kubectl && sudo mv kubectl /bin/kubectl

RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip"
RUN unzip awscli-bundle.zip
RUN sudo ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws
RUN rm ./awscli-bundle.zip
RUN rm -rf ./awscli-bundle/

RUN curl -o aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/1.14.6/2019-08-22/bin/linux/amd64/aws-iam-authenticator
RUN chmod +x ./aws-iam-authenticator
RUN sudo mv ./aws-iam-authenticator /usr/local/bin

COPY crontab.txt crontab.txt
RUN sudo chmod 0644 /home/control/crontab.txt
RUN sudo crontab -u control /home/control/crontab.txt

COPY bin/ /usr/local/bin
run sudo chmod 0777 /usr/local/bin/gen-env.py
COPY test-scripts/ /usr/local/bin
RUN mkdir tenant
COPY tenant tenant
RUN sudo chown -R control:control tenant
RUN mkdir src
COPY test/ src/test
RUN sudo chown -R control:control src
RUN mkdir config
COPY config config
RUN sudo chown -R control:control config
RUN mkdir data
COPY data data
RUN sudo chown -R control:control data
COPY entry.sh /usr/local/bin
RUN sudo chmod 0777 /usr/local/bin/entry.sh
EXPOSE 1323
ENTRYPOINT ["/usr/local/bin/entry.sh"]