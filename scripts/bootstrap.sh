#!/bin/bash
# ./makecert.sh certs joe@random.com
#               ^ dir for certs
#                     ^ email
scripts/metrics.sh &
scripts/tracing.sh &
go run cmd/main.go -config node.yml &
PID1=$!
go run cmd/main.go -config node2.yml &
PID2=$!
go run cmd/main.go -config node3.yml &
PID3=$!
trap "echo \"interrupted\"; sudo kill -9 ${PID1} ${PID2} ${PID3}" SIGHUP SIGINT SIGTERM