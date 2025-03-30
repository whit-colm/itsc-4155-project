#!/bin/sh

# we have to do this to force Postgres to apply these very special
# changes to the postgres db, rather than the default one.
# we do this by sinning :)

read -r -d '' sql << EOM
CREATE EXTENSION pg_cron;
SELECT cron.schedule_in_database(
    'blob-cache-clean-expired', 
    '*/15 * * * *', 
    'CALL blob_cache_clean_expired()', 
    current_database()
);
EOM
echo "${sql}" |  psql -U "${POSTGRES_USER}" -d postgres -f -
