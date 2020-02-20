#!/usr/bin/env bash
docker stop $(docker ps -q --filter ancestor=prom/prometheus)
docker stop $(docker ps -q --filter ancestor=jaegertracing/all-in-one:latest)
lsof -i tcp:2000 | awk 'NR!=1 {print $2}' | xargs kill
lsof -i tcp:2001 | awk 'NR!=1 {print $2}' | xargs kill
lsof -i tcp:2002 | awk 'NR!=1 {print $2}' | xargs kill