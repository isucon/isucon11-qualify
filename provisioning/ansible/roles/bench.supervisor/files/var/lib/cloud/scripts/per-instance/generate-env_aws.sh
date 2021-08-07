#!/usr/bin/env bash

cat << _EOF_ >> /home/isucon/isuxportal-supervisor.env
JIA_SERVICE_URL="https://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4):5000"
_EOF_
chown isucon: /home/isucon/isuxportal-supervisor.env
