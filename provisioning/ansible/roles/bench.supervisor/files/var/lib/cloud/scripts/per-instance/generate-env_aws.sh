#!/usr/bin/env bash

backoff_cnt=10
function curl_with_backoff() {
  local url=$1
  local cnt=0
  while :; do
    result="$(curl -f -m 5 -sL $url)"
    if [ $? -eq 0 ]; then
      break
    else
      if [ $cnt -ge $backoff_cnt ]; then exit 1; fi
      cnt=$(expr $cnt "+" 1)
      sleep $(expr $cnt "*" $cnt)
    fi
  done
  echo "$result"
  unset result
}

cat << _EOF_ >> /home/isucon/isuxportal-supervisor.env
JIA_SERVICE_URL="http://$(curl_with_backoff http://169.254.169.254/latest/meta-data/public-ipv4):5000"
_EOF_
chown isucon: /home/isucon/isuxportal-supervisor.env
