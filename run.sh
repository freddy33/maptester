#!/bin/bash

go build

if [ "$1" == "all" ]; then
    ./maptester regen
    ./maptester test
else
    ./maptester $@
fi
