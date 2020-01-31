#!/usr/bin/env bash

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

[ "$OSTYPE" == "linux-gnu" ]

TARGET_GPP="$1"
TARGET_UID="$2"
TARGET_PATH="$3"
TARGET_CMD="$4"

chmod 0777 "$TARGET_GPP"
env | egrep '^(PATH|docker_tarball|SSH_AUTH_SOCK|GOPATH|GOCACHE|GOPRIVATE|CGO_LDFLAGS_ALLOW|GOLANG_VERSION|BIN|VERSION|STATIC_IMAGE)=' | sed -e 's/^/export /' | tee /tmp/env
useradd --uid "$TARGET_UID" --gid 0 --shell /bin/bash --home-dir /tmp/somebody --create-home somebody
su -l -c '
set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

. /tmp/env
cd '"$TARGET_PATH"'
'"$TARGET_CMD"'
' somebody
