#!/usr/bin/env bash

cd "$(dirname "$(readlink -f "$0")")"
version="$(git describe)"

while read GOOS GOARCH; do
  export GOOS GOARCH
  EXT=""
  [ "$GOOS" == "windows" ] && EXT=".exe"
  go build \
    -ldflags "-X main.version=$version" \
    -o "build/serve_${GOOS}_${GOARCH}${EXT}"
done << EOF
  windows amd64
  darwin  amd64
  linux   amd64
  linux   arm
EOF
