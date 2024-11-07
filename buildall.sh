#!/bin/bash -eu
if [[ -d "bin" ]]; then
  rm -rf bin
fi

./build.sh windows amd64
./build.sh linux amd64
./build.sh darwin amd64
./build.sh windows arm64
./build.sh linux arm64
./build.sh darwin arm64
