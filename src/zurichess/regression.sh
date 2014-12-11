#!/bin/sh

cd `dirname $0`

go build zurichess || exit 1

cutechess-cli \
        -concurrency 4 \
        -games 1000 -pgnout regression.pgn \
        -fcp cmd=`pwd`/zurichess name=zurichess \
        -scp cmd=`pwd`/zurichess name=zurichess \
        -both tc=1/40 proto=uci
