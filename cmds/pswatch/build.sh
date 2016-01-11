#! /bin/sh
#
# build.sh
# Copyright (C) 2016 hzsunshx <hzsunshx@onlinegame-14-51>
#
# Distributed under terms of the MIT license.
#

set -eu

VERSION="0.0.1"
BUILD_DATE=$(date +%Y-%m-%d_%H:%M)

go build -ldflags "-X main.VERSION=$VERSION -X main.BUILD_DATE=$BUILD_DATE" "$@"
