#!/bin/sh

version=`git describe --always --tag`
branch=`git rev-parse --abbrev-ref HEAD`
when=`date +"%F %T"`

go build -ldflags "-X main.buildVersion '$branch-$version' -X main.buildTime '$when'" . || exit 1
cp zurichess zurichess-$branch
cp zurichess zurichess-$branch-$version

echo "Built zurichess-$branch-$version at $when"