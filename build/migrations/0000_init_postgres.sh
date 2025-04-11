#!/bin/sh

# we have to do this to force Postgres to apply these very special
# changes to the postgres db, rather than the default one.
# we do this by sinning :)

cat << EOM | psql -U "${POSTGRES_USER}" -d postgres -f -
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pg_cron;
SELECT cron.schedule_in_database(
    'blob-cache-clean-expired', 
    '*/15 * * * *', 
    'CALL blob_cache_clean_expired()', 
    '${POSTGRES_USER}'
);
EOM