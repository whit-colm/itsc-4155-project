#!/bin/zsh

# Translated by DeepSeek DeepThink-R1 from original bash script

# Explicit PATH and umask (zsh compatible)
export PATH=/usr/bin:/bin:/usr/sbin:/sbin
umask 0077

ME=$(basename "${0}")
ABS_DIR=$(dirname "$(realpath "${0}")")
exec 3>&2 >> "${ABS_DIR}/${ME}-$(date +%s).log" 2>&1
# Set zsh options similar to Bash
set -euvx
uname -a
date

failure () { echo "FAILED." >&3; exit 1 }
trap failure EXIT

# ... (variable definitions remain the same)

print_usage() {
    echo "Debug script for running the software" >&3
    echo "DO NOT USE THIS SCRIPT FOR DEPLOYMENT" >&3
    echo >&3
    echo "Flags:" >&3
    echo "\t-d -\tEnable demo values in DB" >&3
    echo "\t-t -\tAuto-delete pods when done" >&3
    echo "\t-f: [path]\tPath for persistent Postgres volume" >&3
    echo "\t-a: [port]\tPostgres port (default 54321)" >&3
    echo "\t-p: [port]\tApp port (default 8080)" >&3
    exit 2
}

# getopts parsing (zsh compatible)
while getopts 'dtf:a:p:' flag; do
    case "${flag}" in
        d) demo_values='true' ;;
        t) tidy_when_done='true' ;;
        f) persistence="${OPTARG}" ;;
        a) app_port="${OPTARG}" ;;
        p) db_port="${OPTARG}" ;;
        *) print_usage ;;
    esac
done

# ... (directory change and git command remain the same)

jaws_docker_tag=$(date +%Y%m%d%H%M%S)
docker build --target webnative -t jaws:${jaws_docker_tag} .

psql_passwd=$(openssl rand -base64 48)

# ... (flag construction logic remains the same)

docker run --name jaws-psql ${db_flags} -e POSTGRES_PASSWORD=${psql_passwd} -d postgres:17-alpine
docker run --name jaws-app ${app_flags} -e PG_PASSWORD=${psql_passwd} -d jaws:${jaws_docker_tag}

trap - EXIT