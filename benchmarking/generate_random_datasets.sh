#!/usr/bin/env bash
set -euo pipefail

numfiles=${1:-10}
filesize_mbs=${2:-500}

echo $numfiles $filesize_mbs

outputdir=datasets/${numfiles}_files_${filesize_mbs}_mbs
rm -rf $outputdir
mkdir $outputdir

for ((file=1;file<=$numfiles;file++)); do
        head -c ${filesize_mbs}M /dev/urandom > $outputdir/$file.bin
done
