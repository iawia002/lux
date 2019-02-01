#!/bin/bash

set -e
#set -x # debug/verbose execution

# Ensures removal of file
trap 'rm -f profile.out downloader/*.{mp4,download,jpg}' SIGINT EXIT

# go -race needs CGO_ENABLED to proceed
export CGO_ENABLED=1;

for d in $(go list ./... | grep -v vendor); do
	go test -v -race -coverprofile=profile.out -covermode=atomic "${d}"
	if [ -f profile.out ]; then
		cat profile.out > coverage.txt
		rm profile.out
	fi
done
