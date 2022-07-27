#!/bin/bash -ex

go build -o pping ./src/main
cp -f ./pping /usr/local/bin
sudo chown root:staff /usr/local/bin/pping
sudo chmod u+s /usr/local/bin/pping
/usr/local/bin/pping $@

