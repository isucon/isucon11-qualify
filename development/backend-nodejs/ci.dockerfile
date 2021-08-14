FROM node:14.17.4-buster

WORKDIR /webapp/nodejs

#install mariadb-client
RUN apt-get update \
    && apt-get install -y default-mysql-client

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY webapp/nodejs/package.json webapp/nodejs/package-lock.json ./
COPY webapp/nodejs/tsconfig.json ./
RUN npm ci

COPY webapp/nodejs/ .
COPY webapp/public /webapp/public
RUN npm run build

ENTRYPOINT ["dockerize", "-wait=tcp://mysql-backend:3306", "-timeout=60s"]
CMD ["npm", "start"]