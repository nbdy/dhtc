#!/bin/bash

CONTAINER_NAME=dhtc

if [[ $(docker ps -a --filter="name=$CONTAINER_NAME" --filter "status=exited" | grep -w "$CONTAINER_NAME") ]]; then
    echo "Starting container $CONTAINER_NAME"
    docker start $CONTAINER_NAME
elif [[ $(docker ps -a --filter="name=$CONTAINER_NAME" --filter "status=running" | grep -w "$CONTAINER_NAME") ]]; then
    echo "Container is already running"
else
    echo "Container does not exist. Creating it."
    docker build --label $CONTAINER_NAME -t $CONTAINER_NAME .
    echo "Created container. Starting it."
    docker run --name $CONTAINER_NAME -p 4200:4200 -d $CONTAINER_NAME:latest
    docker start $CONTAINER_NAME
fi

echo "Done."
