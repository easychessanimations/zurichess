#!/bin/sh

go build . || exit 1

pgnout=test.$1.vs.$2.pgn
mv -f $pgnout $pgnout~

cutechess-cli \
        -concurrency 4 \
        -games 100 -pgnout $pgnout \
        -fcp cmd=`pwd`/$1 name=$1 \
        -scp cmd=`pwd`/$2 name=$2 \
        -both tc=1/40 proto=uci