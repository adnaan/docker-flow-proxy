#!/bin/bash
if [ "$#" -ne 2 ] ; then
  echo "Usage: ./remove_image SERVICE_NAME BRANCH"
  exit 1
fi

echo "docker rmi  $1:$2"
docker rmi  $1:$2
