FROM python:3.7-alpine
MAINTAINER Kodjo Baah<kafriyie01@bbc.co.uk>

USER root
RUN apk update && apk add curl git jq bash openssh openssh-client openrc rsyslog rsync
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

COPY etc /etc
USER 1000
WORKDIR /home/control

RUN curl --silent --location "https://github.com/weaveworks/eksctl/releases/download/latest_release/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
RUN  sudo mv /tmp/eksctl /usr/local/bin
RUN sudo rc-update add sshd
#RUN sudo touch /run/openrc/softlevel
RUN sudo rc-update add rsyslog boot
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.1/bin/linux/amd64/kubectl
RUN sudo chmod u+x kubectl && sudo mv kubectl /bin/kubectl

RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip"
RUN  unzip awscli-bundle.zip
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
RUN mkdir tennant
COPY tennant tennant
RUN sudo chown -R control:control tennant
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
ENTRYPOINT ["/usr/local/bin/entry.sh"]