#!/bin/bash

if [ ! -d "venv" ] 
then
    rm -rf venv
    python3 --version
    python3 -m venv venv
    source ./venv/bin/activate 
    pip3 install --upgrade pip 
    pip3 install -r requirements.txt 
fi

