#!/bin/bash -eu
if [[ -d "bin" ]]; then
  rm -rf bin
fi

./build.sh linux amd64
./build.sh darwin arm64
