#! /bin/sh

export APP=/app/server/av-server
export ARCH="$(arch)"


if [ "$ARCH" != "x86_64" ]; then
    export APP=/app/server/av-server-$ARCH
fi

echo "Arch: $ARCH [$(uname -a)]"
$APP --data-dir /app/data --frontend-dir /app/flutter/ --config-dir /app/config