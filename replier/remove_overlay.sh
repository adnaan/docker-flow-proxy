#!/bin/bash
if [ "$#" -ne 1 ] ; then
  echo "Usage: ./remove_overlay.sh name"
  exit 1
fi

docker network rm $1
