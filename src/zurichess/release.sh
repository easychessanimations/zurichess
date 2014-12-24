#!/bin/sh

version=`git describe --always --tag`
branch=`git rev-parse --abbrev-ref HEAD`
when=`date +"%F %T"`

go build -ldflags "-X main.buildVersion '$branch-$version' -X main.buildTime '$when'" . || exit 1
cp zurichess zc-$branch
cp zurichess zc-$branch-$version

echo "Built zc-$branch-$version at $when"
