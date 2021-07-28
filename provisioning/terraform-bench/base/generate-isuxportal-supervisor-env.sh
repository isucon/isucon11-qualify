#!/usr/bin/env bash

cat << _EOF_ >> /home/isucon/isuxportal-supervisor.env
ISUXPORTAL_SUPERVISOR_ENDPOINT_URL=${isuxportal_supervisor_endpoint_url}
ISUXPORTAL_SUPERVISOR_TOKEN=${isuxportal_supervisor_token}
ISUXPORTAL_SUPERVISOR_TEAM_ID=${isuxportal_supervisor_team_id}
_EOF_
chown isucon: /home/isucon/isuxportal-supervisor.env
