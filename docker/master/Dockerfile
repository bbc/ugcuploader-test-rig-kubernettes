FROM 546933502184.dkr.ecr.eu-west-2.amazonaws.com/ugcloadtest/jmeter-base:latest
MAINTAINER Kodjo Baah<kafriyie01@bbc.co.uk>

USER root

RUN apk add --no-cache shadow sudo && \
    if [ -z "`getent group 1000`" ]; then \
      addgroup -S -g 1000 jmeter; \
    else \
      groupmod -n jmeter `getent group 1000 | cut -d: -f1`; \
    fi && \
    if [ -z "`getent passwd 1000`" ]; then \
      adduser -S -u 1000 -G jmeter -s /bin/sh jmeter; \
    else \
      usermod -l jmeter -g jmeter -d /home/jmeter -m `getent passwd jmeter | cut -d: -f1`; \
    fi && \
    echo "jmeter ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/jmeter && \
    chmod 0440 /etc/sudoers.d/jmeter

RUN apk add jq

RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip"
RUN  unzip awscli-bundle.zip
RUN ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws

USER 1000
WORKDIR /home/jmeter

ENV TEST_HOME=/home/jmeter/test

RUN mkdir $TEST_HOME
RUN mkdir bin
COPY bin bin
COPY src/test $TEST_HOME
COPY etc /etc
USER root

RUN chmod 0644 /home/jmeter/bin//crontab.txt
RUN crontab -u jmeter /home/jmeter/bin//crontab.txt
RUN chown -R jmeter:jmeter bin
RUN chown jmeter:jmeter $TEST_HOME


USER 1000
ENTRYPOINT ["/home/jmeter/bin/entrypoint.sh"]
EXPOSE 60000
