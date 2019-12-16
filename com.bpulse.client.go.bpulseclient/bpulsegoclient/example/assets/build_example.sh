#!/bin/bash

#set environment variables for Go
export GOROOT=/usr/src/go
export GOBIN=/gobin
export GOPATH=/go

# Install gpm libraries manager
cd /tmp
apt-get install -y git git-core alien pkg-config libaio1
git clone https://github.com/pote/gpm.git && cd gpm;\
    ./configure;\
    make install

# Download code
echo "Download code from repository  "
cd $GOPATH/src
mkdir -p $REPO_NAME_PREFIX
cd $REPO_NAME_PREFIX
git clone $REPO 
ls -la
cd $REPO_NAME

# Move configuration files
cp ./conf/* /conf

# Git checkout to target branch
echo "Changing to branch $BRANCH"
git checkout $BRANCH

# Install third libraries
gpm install

# build
cd com.bpulse.client.go.bpulseclient/bpulsegoclient/example
if [  ! -z BPULSE_BUILD_TAGS ]
then
	echo "Compiling the example tags: $BPULSE_BUILD_TAGS and loop interval $EXAMPLE_LOOP_DELAY_IN_MS"
	go install -ldflags "-X main.LOOP_DELAY_IN_MS $EXAMPLE_LOOP_DELAY_IN_MS" -tags $BPULSE_BUILD_TAGS
else
	echo "Compiling the example without tags and loop interval $EXAMPLE_LOOP_DELAY_IN_MS"
	go install -ldflags "-X main.LOOP_DELAY_IN_MS $EXAMPLE_LOOP_DELAY_IN_MS"
fi
echo "example built."
