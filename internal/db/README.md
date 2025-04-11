# DB Testing Notes

Here's the command to spin 

```
# To run in the root of the repo:
docker run --rm -d --name jaws-psql-test \
    -p 54322:5432 \
    -e POSTGRES_DB=jaws \
    -e POSTGRES_USER=jaws \
    -e POSTGRES_PASSWORD=stinkypassword \
    -v ./build/migrations:/docker-entrypoint-initdb.d \
    postgres:17-alpine
```

Then you can include db tests below, note the `-count 1` flag is necessary to not run cached tests (which don't play normal with the db system).

```
DB_URI='postgresql://jaws:stinkypassword@localhost:54322/jaws' go test ./... -count 1
```

Alternatively, you can connect directly to the container with:

```
docker exec -it jaws-psql-test psql postgresql://jaws:stinkypassword@localhost:5432/jaws
```