#!/bin/sh -e
# ./depth.sh is a script to solve several puzzle test files.

cd `dirname $0`
go build .

./puzzle --quiet --max_depth 4 $* --input mates/mateIn1.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn2.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn3.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn4.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn5.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn6.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn7.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input mates/mateIn8.epd 2> /dev/null

./puzzle --quiet --max_depth 4 $* --input bratko_kopec.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input ecm.epd 2> /dev/null
./puzzle --quiet --max_depth 4 $* --input silent_but_deadly.epd 2> /dev/null
