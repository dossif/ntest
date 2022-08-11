#!/bin/bash -ex

APP="ntest"
go build -o $APP ./src/main
cp -f ./$APP /usr/local/bin
sudo chown root:staff /usr/local/bin/$APP
sudo chmod u+s /usr/local/bin/$APP
