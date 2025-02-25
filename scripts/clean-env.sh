#!/bin/bash

# This script tidies Docker images from your machine, such as those
# generated as a byproduct of creating the JAWS image.
# 
# currently, all it does is remove unnamed "dangling" images and
# all but the most recent-tagged jaws image.

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

# Stop and purge jaws containers
docker kill my-jaws-app || true
docker kill my-jaws-psql || true
docker rm my-jaws-app || true
docker rm my-jaws-psql || true

# remove dangling images and old JAWS images to boot
docker rmi $(docker images -f "dangling=true" -q) || true

docker rmi $(docker images --filter "reference=jaws-app" --format "{{.ID}} {{.CreatedAt}}" \
    | sort -rk 2 \
    | awk 'NR>1 {print $1}') || true

docker rmi $(docker images --filter "reference=jaws-psql" --format "{{.ID}} {{.CreatedAt}}" \
    | sort -rk 2 \
    | awk 'NR>1 {print $1}') || true

# end of script. Therefore the only time failure should be untrapped.
trap - EXIT