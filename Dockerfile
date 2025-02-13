#### Go Backend ####
FROM golang:1.24-alpine as backend
WORKDIR /backend

COPY ./main.go ./
COPY ./pkg ./
COPY ./cmd ./

RUN go build main.go -o /out/backend

#### React Frontend ####
FROM node:23-alpine as frontend

# Set workdir after copy
COPY ./website/ ./
WORKDIR /website

RUN npm install ; \
    npm run build

#### Postgres Database ####
# This 