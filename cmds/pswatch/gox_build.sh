#! /bin/sh
#
# gox_build.sh
# Copyright (C) 2015 hzsunshx <hzsunshx@onlinegame-14-51>
#
# Distributed under terms of the MIT license.
#


mkdir -p release/{armeabi-v7a,x86}

echo "x86 ..."
GOOS=linux GOARCH=386 go build -o release/x86/pswatch

echo "arm ..."
GOOS=linux GOARCH=arm go build -o release/armeabi-v7a/pswatch
