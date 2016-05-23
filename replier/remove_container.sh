#!/bin/bash
if [ "$#" -ne 2 ] ; then
  echo "Usage: ./remove_containers  SERVICE_NAME BRANCH"
  exit 1
fi

docker rm -f $1'-'$2
