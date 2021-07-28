#!/usr/bin/env bash

cat << _EOF_ > /home/isucon/env.sh
MYSQL_HOST="127.0.0.1"
MYSQL_PORT=3306
MYSQL_USER=isucon
MYSQL_DBNAME=isucondition
MYSQL_PASS=isucon
SERVER_PUBLIC_ADDRESS="$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)"
SERVER_PUBLIC_PORT=80
_EOF_
chown isucon: /home/isucon/env.sh
