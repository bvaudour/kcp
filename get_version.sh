#!/usr/bin/env bash
cat src/kcp/kcp.go | grep "VERSION " | grep -v "D_VERSION" | sed 's/.*"\(.*\)".*/\1/'
