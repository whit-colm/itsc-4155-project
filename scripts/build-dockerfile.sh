#!/bin/bash

# This script wraps the kinda loathsome set of tasks involved in making
# this project work. In addition to building the actual app's container
# We also have to get it running and deploy/migrate/whatever a Postgres
# container along with it.
# There's also a TBD ordeal regarding pre-populating the DB with dummy
# values for testing purposes. This is one of those comments that you
# shouldn't be reading more than a week after it was initially
# committed to the repo.

# explicit PATH, and umask
export PATH=/usr/bin:/bin/:/usr/sbin:/sbin
umask 0077

ME=$(basename "${0}")
ABS_DIR=$(dirname "$(realpath "${0}")")
exec 3>&2 >> "${ABS_DIR}"/"${ME}"-$(date +%s).log 2>&1
# sets:
# - errexit: exit if simple command fails (nonzero return value)
# - nounset: write message to stderr if trying to expand unset variable
# - verbose: Write input to standard error as it is read.
# - xtrace : Write to stderr trace for each command after expansion.
set -euvx
uname -a
date

failure () { echo FAILED. >&3; exit 1; }
trap failure EXIT

# move to root of repo. This does assume the script is "somewhere" in
# the repo. Which I think is a perfectly healthy assumption to make.
cd "${ABS_DIR}"
cd "$(git rev-parse --show-toplevel)"

# Get the time in YYYYMMDDhhmmss and set it as the docker tag for this build
# (i.e. 2025-02-13 22:42:51 should be rendered as 20250213224251)
readonly jaws_docker_tag =  $(date +%Y%m%d%H%M%S)
docker build --target webnative -t ${jaws_docker_tag}

# randomly generate a password for PostgreSQL
readonly psql_passwd = $(openssl rand -base64 48)

# We have to start up the Postgres container, which is a pain in the--
docker run --name jaws-psql -e POSTGRES_PASSWORD=${psql_passwd} -d postgres:17-alpine
docker run --name jaws-app -e POSTGRES_PASSWORD=${psql_passwd} -d jaws:${jaws_docker_tag}


# end of script. Therefore the only time failure should be untrapped.
trap - EXIT