#!/usr/bin/env bash
docker run --rm \
-p 9090:9090 \
-p 9100:9100 \
-v /Users/a1/GolandProjects/rounds/prometheus.yml:/etc/prometheus/prometheus.yml \
prom/prometheus

