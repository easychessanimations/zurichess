#!/bin/sh

if [ -z $1 ] || [ -z $2 ]; then
        echo "Usage: $0 release1 release2"
        exit 1
fi

pgnout=debug$1.vs.$2.pgn
mv -f $pgnout $pgnout~
echo "Will write game in $pgnout"

cutechess-cli \
        -debug \
        -srand $RANDOM \
        -repeat \
        -recover \
        -games 1 \
        -concurrency 1 \
        -resign movecount=3 score=400 \
        -draw movenumber=34 movecount=8 score=20 \
        -openings file=2moves_v1.pgn format=pgn order=random \
        -pgnout $pgnout \
        -engine cmd=`pwd`/$1 name=$1 whitepov \
        -engine cmd=`pwd`/$2 name=$2 whitepov \
        -each tc=40/60 proto=uci
