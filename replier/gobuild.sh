#!/bin/bash
if [ "$#" -ne 3 ] || ! [[ "$1" =~ ^[0-9]+$ ]]; then
  echo "Usage: ./build.sh PORT SERVICE_NAME BRANCH"
  exit 1
fi

#env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.port=$1 -X main.name=$2 -X main.branch=$3" -o reply .
go build -ldflags "-X main.port=$1 -X main.name=$2 -X main.branch=$3" -o reply .
