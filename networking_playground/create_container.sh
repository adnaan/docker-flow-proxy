#!/bin/bash
if [ "$#" -ne 4 ] || ! [[ "$1" =~ ^[0-9]+$ ]]; then
  echo "Usage: ./create_container PORT SERVICE_NAME BRANCH OVERLAY"
  exit 1
fi

echo "Service "$2-$3
rm -f reply entry.sh
echo '#!/bin/sh' >> entry.sh
echo 'socat TCP-LISTEN:80,fork TCP:localhost:'$1' &' >> entry.sh
echo './reply' >> entry.sh
echo "Build Go!"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags \
 "-s -X main.port=$1 -X main.name=$2 -X main.branch=$3" -o reply .

echo "Build Image"
docker build --rm=true --build-arg PORT=$1 -t service:$2-$3 .

echo "Run Container"
if [ $4 = "true" ]; then
  docker run -d --net=$3 --net-alias=$2.myntra.com --name=$2-$3 service:$2-$3
else
  docker run -d --name=$2-$3 service:$2-$3
fi
