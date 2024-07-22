#!/usr/bin/env bash

VERSION=$(git describe --tags)

function build() {
    printf "Compiling version ${VERSION} for ${1}/${2}\t"
    GOOS=${1} GOARCH=${2} go build -ldflags "-X 'main.Version=${VERSION}'" -o dist/wid-notifier_${VERSION}_${1}_${3}
    echo "✅"
}

build linux "386" i386
build linux amd64 amd64
build linux arm arm
build linux arm64 arm64
