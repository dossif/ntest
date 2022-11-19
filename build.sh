#!/bin/bash -ex

APP="ntest"
go build -v -a -trimpath -o $APP ./cmd/ntest
cp -f ./$APP /usr/local/bin
gsudo chown root:staff /usr/local/bin/$APP
gsudo chmod u+s /usr/local/bin/$APP
rm ./ntest
