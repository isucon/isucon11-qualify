FROM node:15.12 as frontend
WORKDIR /app

COPY webapp/frontend/package*.json ./
RUN npm ci

COPY webapp/frontend .
RUN npm run build


FROM golang:1.16.5-buster

WORKDIR /development
COPY development/backend-go/air.toml .

#install mariadb-client
RUN apt-get update \
    && apt-get install -y default-mysql-client

WORKDIR /webapp/go

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && go get -u github.com/cosmtrek/air

COPY webapp/go/go.mod .
COPY webapp/go/go.sum .

RUN go mod download

COPY --from=frontend /public /webapp/public
