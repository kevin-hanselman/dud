#!/usr/bin/env bash

echo "Running dvc"
docker run --rm -v $PWD/benchmarking/datasets:/datasets duc:benchmark /src/benchmarking/run_dvc.sh $1
cat benchmarking/datasets/$1_dvc.txt

echo "Running duc"
docker run --rm -v $PWD/benchmarking/datasets:/datasets duc:benchmark /src/benchmarking/run_duc.sh $1
cat benchmarking/datasets/$1_duc.txt
