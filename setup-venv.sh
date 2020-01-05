#!/bin/bash

if [ ! -d "venv" ] 
then
    rm -rf venv
    python3 --version
    python3 -m venv venv
    source ./venv/bin/activate 
    pip install --upgrade pip 
    pip install -r requirements.txt 
fi

