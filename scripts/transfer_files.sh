#!/bin/bash
docker run --rm -v $1 $2 cp -r /$3/public /tmp_mount/
