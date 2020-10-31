#!/bin/bash
set -euo pipefail

mkdir fake_remote

dud config set remote fake_remote

dud push
