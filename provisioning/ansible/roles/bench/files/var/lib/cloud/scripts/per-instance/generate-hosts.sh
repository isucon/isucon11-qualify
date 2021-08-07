#!/usr/bin/env bash

cat << _EOF_ >> /etc/hosts
$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
JIA_SERVICE_URL="https://:5000"
_EOF_
chown isucon: /home/isucon/isuxportal-supervisor.env
