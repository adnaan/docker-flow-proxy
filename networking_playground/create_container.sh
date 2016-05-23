#!/bin/bash
if [ "$#" -ne 4 ] || ! [[ "$1" =~ ^[0-9]+$ ]]; then
  echo "Usage: ./create_container PORT SERVICE_NAME BRANCH OVERLAY"
  exit 1
fi

rm reply
echo "Service "$2:$3
echo "Build Go!"
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.port=$1 -X main.name=$2 -X main.branch=$3" -o reply .
echo "Build Image"
docker build --rm=true --build-arg PORT=$1 -t $2:$3 .

echo "Run Container"
if [ $4 = "true" ]; then
  docker run -d --net=$3 --net-alias=$2.myntra.com --name=$2'-'$3 $2:$3
else
  docker run -d --name=$2'-'$3 $2:$3
fi
