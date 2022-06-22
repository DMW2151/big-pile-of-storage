#! /bin/bash

# User Data for the Client Instance - Give the Instance Python3 w. some AWS utils for
# ease of devlopment/debugging

sudo apt-get update -y &&\
sudo apt-get install python3-pip -y &&\
python3 -m pip install boto3 && 
echo "Done"
