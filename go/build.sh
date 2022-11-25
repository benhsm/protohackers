#!/bin/sh
# Usage: build.sh <challenge directory> <executable_name>
cd "$1" || exit 1
# CGO will sometimes use dynamic linking and introduce dependencies on shared
# libraries, but here I want static binaries that I can just scp to a server
# and run.
CGO_ENABLED=0 GOOS=linux go build -a -o "$2"
