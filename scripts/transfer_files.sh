#!/bin/bash
docker run -v $1 $2 cp -r /$3/public /tmp_mount/
