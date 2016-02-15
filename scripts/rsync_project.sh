#!/bin/bash

fileToTransfer=$1
remoteLocation=$2

# First we will attempt to make the remote directory
echo "Creating directory $remoteLocation"
ssh -i /home/franklin/.ssh/id_rsa franklin@islstatic.com "mkdir -p $remoteLocation"

# Than we will copy the files
echo "Transfering files to $remoteLocation"
rsync -azIe "ssh -i /home/franklin/.ssh/id_rsa" $fileToTransfer franklin@islstatic.com:$remoteLocation
