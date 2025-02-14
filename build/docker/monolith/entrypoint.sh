#!/bin/sh

umask 0077
set -euvx

exec docker-entrypoint.sh &

nginx &

/app/myapp &

wait -n

exit $?