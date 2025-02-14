# Use multi-stage builds to keep the final image small

#### Go Backend ####
FROM golang:1.24-alpine as backend
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY pkg ./pkg
COPY cmd ./cmd

RUN go build -o jaws main.go

#### React Frontend ####
FROM node:23-alpine as frontend

WORKDIR /website
COPY website/ ./

RUN npm install ; \
    npm run build

#### Postgres Database ####
## A webnative deployment should have a database external to itself.
## This is useful for testing and debugging purposes - because of how
## complex postgres is, we use it as a base for the monolith
FROM nginx:1-alpine as webnative

# copy outputs compiled in prior steps
COPY --from=backend /app/jaws /app/jaws
COPY --from=frontend /website/build /var/www/html
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
ENV DEBUG_MODE=false

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
    CMD curl -f http://localhost:8080/health

# We do not set ENTRYPOINT or CMD; the default one with nginx works.