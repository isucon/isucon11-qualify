#!/bin/bash
set -x

retried=0
while [[ $retried -le 15 ]]; do
  sleep $(( RANDOM % 15 ))
  /opt/isucon-env-checker/isucon-env-checker && exit 0
  retried=$(( retried + 1 ))
done
