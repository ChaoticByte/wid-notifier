#!/usr/bin/env bash

VERSION=$(git describe --tags)

# i386
GOOS=linux GOARCH=386 go build -o dist/wid-notifier_${VERSION}_linux_i386

# amd64
GOOS=linux GOARCH=amd64 go build -o dist/wid-notifier_${VERSION}_linux_amd64

# arm
GOOS=linux GOARCH=arm go build -o dist/wid-notifier_${VERSION}_linux_arm

# arm64
GOOS=linux GOARCH=arm64 go build -o dist/wid-notifier_${VERSION}_linux_arm64
