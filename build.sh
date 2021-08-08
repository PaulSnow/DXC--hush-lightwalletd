#!/bin/bash
# Copyright 2021 The Hush Developers
# Released under GPLv3

# Check if go is installed on system and exits if it is not
if ! [ -x "$(command -v go)" ]; then
  echo 'Error: go is not installed. Install go and try again.' >&2
  exit 1
fi

# now to compiling...
cd `pwd`/cmd/server
go build main.go
# move compiled main.go to lightwalletd
mv main `pwd`/../../lightwalletd
echo "lightwalletd is now compiled for you"
