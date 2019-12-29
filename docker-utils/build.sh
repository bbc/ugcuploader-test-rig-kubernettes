#!/bin/bash

JMETER_VERSION="5.1.1"

# Example build line
# --build-arg IMAGE_TIMEZONE="Europe/Amsterdam"
docker build  --build-arg JMETER_VERSION=${JMETER_VERSION} -t "ugcupload/jmeter:${JMETER_VERSION}" .
