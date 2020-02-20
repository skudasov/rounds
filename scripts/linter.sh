#!/usr/bin/env bash
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.21.0 golangci-lint run -v
