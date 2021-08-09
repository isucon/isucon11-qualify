FROM node:16.6.1 as frontend
WORKDIR /app

COPY webapp/frontend/package*.json ./
RUN npm ci

COPY webapp/frontend .
RUN npm run build

FROM rust:1.54.0-buster

WORKDIR /webapp/rust

#install mariadb-client
RUN apt-get update \
    && apt-get install -y default-mysql-client

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY webapp/rust/Cargo.lock webapp/rust/Cargo.toml ./
RUN mkdir src &&  echo 'fn main() {}' > src/main.rs && cargo build --locked && rm src/main.rs target/debug/deps/isucondition-*

COPY webapp/rust/ ./
RUN cargo build --locked --frozen

COPY --from=frontend /public /webapp/public

ENTRYPOINT ["dockerize", "-wait=tcp://mysql-backend:3306", "-timeout=60s", "./target/debug/isucondition"]

# vim: set ft=dockerfile:
