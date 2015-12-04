#!/bin/bash

fileToTransfer=$1
remoteLocation=$2

ssh -i /home/franklin/.ssh/id_rsa franklin@islstatic.com mkdir -p $remoteLocation && \
  rsync -azIe "ssh -i /home/franklin/.ssh/id_rsa" $fileToTransfer franklin@islstatic.com:$remoteLocation
