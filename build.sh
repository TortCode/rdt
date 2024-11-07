#!/bin/bash -eu

GOOS=$1 GOARCH=$2 go build -o "bin/$1/$2/server" cmd/server/main.go && echo "BUILD: server TARGET: $1/$2"
GOOS=$1 GOARCH=$2 go build -o "bin/$1/$2/client" cmd/client/main.go && echo "BUILD: client TARGET: $1/$2"