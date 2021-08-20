#!/usr/bin/env bash

TARGET_FILE=/home/isucon/isuxportal-supervisor.env
PUBLIC_ADDR="$(curl -s --retry 5 --retry-connrefused --max-time 10 --connect-timeout 5 http://169.254.169.254/latest/meta-data/public-ipv4)"

if egrep '^JIA_SERVICE_URL' ${TARGET_FILE} &> /dev/null; then
perl -pi -e 's|^(JIA_SERVICE_URL=).*$|$1"http://'${PUBLIC_ADDR}':5000"|g' ${TARGET_FILE}
else
cat << _EOF_ >> ${TARGET_FILE}
JIA_SERVICE_URL="http://${PUBLIC_ADDR}:5000"
_EOF_
fi

chown isucon: ${TARGET_FILE}
