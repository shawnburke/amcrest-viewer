#! /bin/sh

export APP=/app/server/av-server

[[ $(uname -p) != "x86_64" ]] && export APP=/app/server/av-server-$(uname -p)


$APP --data-dir /app/data --frontend-dir /app/flutter/ --config-dir /app/config