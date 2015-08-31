#!/bin/bash

echo "destroying temporary docker container and image"

image_id=$(docker images | grep "<none>" | awk '{print $3}')

# stop container
docker ps -a | grep "$image_id" | awk '{print $1}' | xargs docker stop
# delete container
docker ps -a | grep "$image_id" | awk '{print $1}' | xargs docker rm
# delete image
echo $image_id | xargs docker rmi
