#! /bin/sh
#
# gox_build.sh
# Copyright (C) 2015 hzsunshx <hzsunshx@onlinegame-14-51>
#
# Distributed under terms of the MIT license.
#


mkdir -p release/{armeabi-v7a,x86}

echo "x86 ..."
GOOS=linux GOARCH=386 ./build.sh -o release/x86/pswatch

echo "arm ..."
GOOS=linux GOARCH=arm ./build.sh -o release/armeabi-v7a/pswatch

echo "arm64 ..."
GOOS=linux GOARCH=arm64 ./build.sh -o release/arm64-v8a/pswatch
