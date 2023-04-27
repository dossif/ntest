#!/bin/bash -ex

APP="ntest"
VER="0.0.1"
go build -v -a -trimpath -ldflags "-X main.appVersion=${VER}" -o $APP ./cmd/ntest
cp -f ./$APP /usr/local/bin
gsudo chown root:staff /usr/local/bin/$APP
gsudo chmod u+s /usr/local/bin/$APP
rm ./ntest
