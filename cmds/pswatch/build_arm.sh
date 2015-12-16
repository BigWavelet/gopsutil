#! /bin/sh
#
# build_arm.sh
# Copyright (C) 2015 hzsunshx <hzsunshx@onlinegame-14-51>
#
# Distributed under terms of the MIT license.
#


set -e
GOOS=linux GOARCH=arm go build
echo "Push ..."
adb -P ${PORT:-"13020"} push pswatch  /data/local/tmp/ 
