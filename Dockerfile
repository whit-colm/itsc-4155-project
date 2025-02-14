# Use multi-stage builds to keep the final image small


#### Go Backend ####
FROM golang:1.24-alpine as backend
WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download


COPY main.go ./
COPY pkg ./
COPY cmd ./


RUN go build -o /jaws main.go


#### React Frontend ####
FROM node:23-alpine as frontend


# Set workdir after copy
COPY website/ ./
WORKDIR /website


RUN npm install ; \
    npm run build


#### Postgres Database ####
## A webnative deployment should have a database external to itself.
## This is useful for testing and debugging purposes - because of how
## complex postgres is, we use it as a base for the monolith
FROM nginx:1-alpine as webnative


# copy outputs compiled in prior steps
COPY --from=backend /out/backend /app/jaws
COPY --from=frontend /website/build /var/www/html
# copy build configs to appropriate places in container
COPY build/docker/nginx.conf /etc/nginx/nginx.conf
COPY build/docker/monolith/initdb.sql /docker-entrypoint-initdb.d/initdb.sql
# copy custom entrypoint and override default one used by psql
COPY build/docker/monolith/entrypoint.sh /entrypoint.sh


RUN chmod +x /entrypoint.sh


ENTRYPOINT [ "/entrypoint.sh" ]
