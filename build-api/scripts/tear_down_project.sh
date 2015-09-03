#!/bin/bash

echo "destroying temporary docker container and image"

# If everything went well, we'll have a named image to destroy
image_id=$(docker images | grep "franklin_builder_tmp" | awk '{print $3}')
if [ -n "$image_id" ]; then
    # stop container
    docker ps -a | grep "$image_id" | awk '{print $1}' | xargs docker stop
    # delete container
    docker ps -a | grep "$image_id" | awk '{print $1}' | xargs docker rm
    # delete image
    echo $image_id | xargs docker rmi
fi

# regardless of above, attempt to find and destroy orphan containers/images
orphan_id=$(docker images | grep "<none>" | awk '{print $3}')
if [ -n "$orphan_id" ]; then
    # stop containers
    docker ps -a | grep "$orphan_id" | awk '{print $1}' | xargs docker stop
    # delete containers
    docker ps -a | grep "$orphan_id" | awk '{print $1}' | xargs docker rm
    # delete images
    echo $orphan_id | xargs docker rmi
fi
