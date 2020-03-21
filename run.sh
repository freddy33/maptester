#!/bin/bash

if [ "$1" == "all" ]; then
    go run runner/main.go regen
    go run runner/main.go test
else
    go run runner/main.go $@
fi
