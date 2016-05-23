#!/bin/bash
if [ "$#" -ne 2 ] ; then
  echo "Usage: ./create_overlay.sh subnet name"
  exit 1
fi

docker network create -d overlay --subnet=$1 $2
