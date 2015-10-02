#!/usr/bin/env bash
#cat src/sysutil/util.go | grep "VERSION " | grep -v "D_VERSION" | sed 's/.*"\(.*\)".*/\1/'
sed -nre 's/.*VERSION.*"(.*)".*/\1/p' src/sysutil/util.go
