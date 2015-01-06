#!/bin/sh
#
# test.sh evaluate two releases of zurichess.
# Requires cutechess-cli (latest version) in PATH.
#
# Usage: ./test.sh release1 release2
#
# 2moves_v1.pgn (not included in the repository) is taken Fishtest.
# Similarly -draw and -resign flags are take from Fishtest.

if [ -z $1 ] || [ -z $2 ]; then
        echo "Usage: $0 release1 release2"
        exit 1
fi

pgnout=test.$1.vs.$2.pgn
mv -f $pgnout $pgnout~
echo "Will write games in $pgnout"

cutechess-cli \
        -srand $RANDOM \
        -repeat \
        -recover \
        -games 5000 \
        -concurrency 2 \
        -resign movecount=8 score=500 \
        -draw movenumber=40 movecount=8 score=20 \
        -openings file=2moves_v1.pgn format=pgn order=random \
        -sprt elo0=1 elo1=5 alpha=0.01 beta=0.01 \
        -pgnout $pgnout \
        -engine cmd=`pwd`/$1 name=$1 whitepov \
        -engine cmd=`pwd`/$2 name=$2 whitepov \
        -each tc=40/15+0.05 proto=uci
