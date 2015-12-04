#!/bin/bash

fileToTransfer=$1
remoteLocation=$2

rsync -azIe "ssh -i /home/franklin/.ssh/id_rsa" $fileToTransfer franklin@islstatic.com:$remoteLocation
