#!/bin/bash -ex

go build -o pping ./src/main
sudo chown root:staff ./pping
sudo chmod u+s ./pping
./pping $1