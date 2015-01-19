#!/bin/sh -e
# ./all.sh is a script to solve several puzzle test files.

cd `dirname $0`
go build .

./puzzle --quiet --deadline 100ms $* --input mates/mateIn1.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn2.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn3.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn4.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn5.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn6.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn7.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input mates/mateIn8.epd 2> /dev/null

./puzzle --quiet --deadline 500ms $* --input bratko_kopec.epd 2> /dev/null
./puzzle --quiet --deadline 100ms $* --input ecm.epd 2> /dev/null
./puzzle --quiet --deadline 500ms $* --input silent_but_deadly.epd 2> /dev/null
