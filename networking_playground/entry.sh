#!/bin/sh
socat TCP-LISTEN:80,fork TCP:localhost:4444 &
./reply
