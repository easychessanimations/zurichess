#!/bin/sh

branch=`git rev-parse --abbrev-ref HEAD`
release=${branch#release.}
version=`git describe --always --tag`
when=`date +"%F %T"`

export GOPATH=`realpath $(dirname $0)/..`

# make goos goarch extension
function make {
        GOOS=$1 GOARCH=$2 go build \
                -a \
                -ldflags "-X main.buildVersion '$branch-$version' -X main.buildTime '$when'" \
                -o zurichess-$release-$1-$2$3 \
                zurichess
}

make   linux amd64 ""
make   linux   386 ""
make windows amd64 ".exe"
make windows   386 ".exe"