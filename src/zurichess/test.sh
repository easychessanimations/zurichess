#!/bin/sh

go build . || exit 1

pgnout=test.$1.vs.$2.pgn
mv -f $pgnout $pgnout~

cutechess-cli \
        -concurrency 2 \
        -games 500 \
        -pgnin 2moves_v1.pgn \
        -pgnout $pgnout \
        -fcp cmd=`pwd`/$1 name=$1 \
        -scp cmd=`pwd`/$2 name=$2 \
        -both tc=40/60 proto=uci
