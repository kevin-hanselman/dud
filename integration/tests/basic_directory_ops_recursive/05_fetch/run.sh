#!/bin/bash
set -euo pipefail

rm -rf .dud/cache/*

dud fetch
