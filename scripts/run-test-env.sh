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

# Define flags which change how we handle the execution of the pod
demo_values=''
tidy_when_done=''
persistence=''
app_port='8080'
db_port='54321'

# Print usage
print_usage() {
    echo -e "Debug script for running the software" >&3
    echo -e "DO NOT USE THIS SCRIPT FOR DEPLOYMENT" >&3
    echo -e >&3
    echo -e "Flags:" >&3
    echo -e "\t-d -\tEnable demo values in DB" >&3
    echo -e "\t-t -\tHave pods auto-delete themselves when done" >&3
    echo -e "\t-f: [path]\tPath to volume to mount for persistent Postgres" >&3
    echo -e "\t-a: [port]\tPostgres port to expose (default 54321)" >&3
    echo -e "\t-p: [path]\tNginx port to expose (default 8080)" >&3
    exit 2
}

# Iterate using getopts to get flags and assign to vars
# From https://stackoverflow.com/a/21128172; licensed CC BY-SA 4.0
while getopts 'dtf:a:p:' flag; do
    case "${flag}" in
        d) 
            demo_values='true'
            ;;
        t)
            tidy_when_done='true'
            ;;
        f) 
            persistence="${OPTARG}" 
            ;;
        a) 
            app_port="${OPTARG}" 
            ;;
        p)
            db_port="${OPTARG}" 
            ;;
        ?) 
            trap print_usage EXIT
            exit
            ;;
    esac
done

# move to root of repo. This does assume the script is "somewhere" in
# the repo. Which I think is a perfectly healthy assumption to make.
cd "${ABS_DIR}"
cd "$(git rev-parse --show-toplevel)"

# Get the time in YYYYMMDDhhmmss and set it as the docker tag for this build
# (i.e. 2025-02-13 22:42:51 should be rendered as 20250213224251)
readonly jaws_docker_tag=$(date +%Y%m%d%H%M%S)
docker build --target webnative -t jaws:${jaws_docker_tag} .

# randomly generate a password for PostgreSQL
readonly psql_passwd=$(openssl rand -base64 48)

# We have to run two containers, the app and the database
# We also have to set runtime flags based on script flags passed.
# we can set some flags already because app_port and db_port are always
# set.
db_flags="-p ${db_port}:5432 "
app_flags="-p ${app_port}:80 "

# check if we should use demo values
# TODO: implement demo values
if [[ -n "${demo_values}" ]]; then
    echo "Sorry, demo values have not been implemented yet."
fi
# Now check if we need to tidy our containers; which is to say we --rm
if [[ -n "${tidy_when_done}" ]]; then
    app_flags+='--rm '
    db_flags+='--rm '
fi
# Finally we add a volume to the db if we must
if [[ -n "${persistence}" ]]; then
    db_flags+="--mount type=bind,source=${persistence},target=/var/lib/postgresql/data "
fi

# Now we can run our containers with the given flags
docker run --name jaws-psql ${db_flags} \
    -e POSTGRES_PASSWORD=${psql_passwd} \
    -e POSTGRES_USER="jaws" \
    -e POSTGRES_DB="jaws" \
    -d postgres:17-alpine

docker run --name jaws-app ${app_flags} \
    -e PG_PASSWORD=${psql_passwd} \
    -e PG_USER="jaws" \
    -e PG_DATABASE="jaws" \
    -e PG_HOST="jaws-psql" \
    -e PG_PORT=${db_port} \
    -d jaws:${jaws_docker_tag}

# end of script. Therefore the only time failure should be untrapped.
trap - EXIT