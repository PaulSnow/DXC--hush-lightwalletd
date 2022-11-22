#!/usr/bin/env bash
# Copyright 2021-2022 Duke Leto and The Hush Developers
# Distributed under the GPLv3 software license, see the accompanying
# file LICENSE or https://www.gnu.org/licenses/gpl-3.0.en.html
# Purpose: Script to build Hush lightwalletd on x86 64-bit arch
# Check if go is installed on system and exits if it is not

set -e

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
echo "You have go installed, so starting to compile Hush lightwalletd for you..."
# cd `pwd`/cmd/server
go build -o lightwalletd main.go
# mv lightwalletd `pwd`/../../lightwalletd
echo ""
echo "Hush lightwalletd is now compiled for you. Enjoy and reach out if you need support."
echo "For options, run ./lightwalletd --help"
