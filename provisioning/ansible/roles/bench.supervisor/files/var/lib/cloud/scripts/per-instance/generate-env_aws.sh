#!/usr/bin/env bash

cat << _EOF_ >> /home/isucon/isuxportal-supervisor.env
JIA_SERVICE_URL="http://$(curl -s --retry 5 --retry-connrefused --max-time 10 --connect-timeout 5 http://169.254.169.254/latest/meta-data/public-ipv4):5000"
_EOF_
chown isucon: /home/isucon/isuxportal-supervisor.env
