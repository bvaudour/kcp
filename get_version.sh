#!/usr/bin/env bash
cat src/kcp/kcp.go | grep "versionNumber " | sed 's/.*"\(.*\)".*/\1/'