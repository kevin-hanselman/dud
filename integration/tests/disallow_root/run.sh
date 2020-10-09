#!/bin/bash

duc init

if sudo duc status; then
    echo 'Expected command to fail' 1>&2
    exit 1
fi
