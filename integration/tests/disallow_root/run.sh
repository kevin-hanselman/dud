#!/bin/bash

dud init

if sudo dud status; then
    echo 'Expected command to fail' 1>&2
    exit 1
fi
