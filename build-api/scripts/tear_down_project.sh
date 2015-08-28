#!/bin/bash

echo "destroying temporary docker container and image"

# Destroy orphans (possible for tmp? Maybe if theres an error?)
# docker images | grep "<none>" | awk '{print $3}' | xargs docker rmi
# Destroy the container
docker ps -a | grep 'tmp_web' | awk '{print $1}' | xargs docker rm
# Destroy the image
docker images | grep "tmp_web" | awk '{print $3}' | xargs docker rmi

# shut down docker (?)
# docker-compose stop
