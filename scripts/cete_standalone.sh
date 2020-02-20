#!/usr/bin/env bash
DOCKER_NAME="cete-node1"
STORAGE_PATH="/tmp/cete/node1"
docker stop ${DOCKER_NAME}
sudo rm -rf ${STORAGE_PATH}
docker run --rm --name ${DOCKER_NAME} \
    -p 5050:5050 \
    -p 6060:6060 \
    -p 8080:8080 \
    mosuka/cete:latest cete start \
      --node-id=node1 \
      --bind-addr=:6060 \
      --grpc-addr=:5050 \
      --http-addr=:8080 \
      --data-dir=/tmp/cete/node1 &
