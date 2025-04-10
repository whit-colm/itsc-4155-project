# Use multi-stage builds to keep the final image small

#### Go Backend ####
FROM golang:1.24-alpine AS pre-backend
WORKDIR /app

COPY go.mod ./ 
COPY go.sum ./ 
COPY main.go ./ 
COPY pkg ./pkg 
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o jaws main.go

#### React Frontend ####
FROM node:23-alpine AS pre-frontend

WORKDIR /website
COPY website/ ./

RUN npm install ; \
    npm run build

#### Postgres Database ####
## A webnative deployment should have a database external to itself.
## This is useful for testing and debugging purposes - because of how
## complex postgres is, we use it as a base for the monolith
FROM nginx:1-alpine AS app

# copy outputs compiled in prior steps
COPY --from=pre-backend /app/jaws /app/jaws
RUN chown nginx:nginx /app/jaws
COPY --from=pre-frontend /website/build /var/www/html
RUN chown -R nginx:nginx /var/www/html/
# copy in backend runner to /docker-entrypoint.d/ (which nginx runs at
# startup) and set appropriate ownership and permission bits. Also copy
# nginx.conf + templates (templates can use ENV vars)
COPY build/docker/start-backend.sh /docker-entrypoint.d/start-backend.sh
RUN chown nginx:nginx /docker-entrypoint.d/start-backend.sh ; \
    chmod +x /docker-entrypoint.d/start-backend.sh
COPY build/docker/nginx.conf /etc/nginx/nginx.conf
RUN chown nginx:nginx /etc/nginx/nginx.conf
COPY build/docker/default.conf.template /etc/nginx/templates/
RUN chown nginx:nginx /etc/nginx/templates/*

# Environment variables for
ENV DEBUG_MODE='false'

# Environment variables for PostgreSQL
ENV PG_HOST=localhost \
    PG_PORT=5432 \
    PG_USER=postgres \
    PG_PASSWORD=secret \
    PG_DB=jaws

# Environment variables for Nginx
ENV NGINX_HOST=localhost

# Establish healthcheck for backend
HEALTHCHECK --interval=30s --timeout=30s --start-period=10s --retries=3 \
    CMD curl -f http://127.0.0.1:9000/api/health

# We do not set ENTRYPOINT or CMD; the default one with nginx works.

FROM paradedb/paradedb AS db

COPY build/migrations /docker-entrypoint-initdb.d
#COPY --from=pre-db /usr/lib/postgresql17/pg_cron.so /usr/local/lib/postgresql/
#COPY --from=pre-db /usr/share/postgresql/extension/pg_cron* /usr/local/share/postgresql/extension
#RUN chown postgres:postgres /docker-entrypoint-initdb.d/* && \
#    chmod +x /docker-entrypoint-initdb.d/*.sh

CMD ["postgres",\
    "-c",\
    "shared_preload_libraries=pg_cron"]