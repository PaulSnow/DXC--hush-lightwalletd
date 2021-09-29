#!/usr/bin/env bash
# Copyright 2021 The Hush Developers
# Released under GPLv3

# Check if go is installed on system and exits if it is not
if ! [ -x "$(command -v go)" ]; then
  echo 'Error: go is not installed. Install go and try again.' >&2
  exit 1
fi

echo ""
echo "Welcome to the Hush magic folks..."
echo ""
echo ".------..------..------..------..------..------..------..------..------..------..------..------."
echo "|L.--. ||I.--. ||G.--. ||H.--. ||T.--. ||W.--. ||A.--. ||L.--. ||L.--. ||E.--. ||T.--. ||D.--. |"
echo "| :/\: || (\/) || :/\: || :/\: || :/\: || :/\: || (\/) || :/\: || :/\: || (\/) || :/\: || :/\: |"
echo "| (__) || :\/: || :\/: || (__) || (__) || :\/: || :\/: || (__) || (__) || :\/: || (__) || (__) |"
echo "| '--'L|| '--'I|| '--'G|| '--'H|| '--'T|| '--'W|| '--'A|| '--'L|| '--'L|| '--'E|| '--'T|| '--'D|"
echo "+------'+------'+------'+------'+------'+------'+------'+------'+------'+------'+------'+------'"

# now to compiling...
echo ""
echo "You have go installed, so starting to compile hush lightwalletd for you..."
cd `pwd`/cmd/server
go build -o lightwalletd main.go
mv lightwalletd `pwd`/../../lightwalletd
echo ""
echo "lightwalletd is now compiled for you."
echo "for options, run ./lightwalletd --help"
