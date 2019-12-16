#!/bin/bash

apt-get install -y curl git mercurial make binutils bison gcc build-essential
curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer > /tmp/gvm_install
chmod +x /tmp/gvm_install
/tmp/gvm_install > /dev/null
source /root/.gvm/scripts/gvm
export PATH=$PATH:~/bin:~/.gvm/bin
gvm install $GO_VERSION
