#!/usr/bin/env bash
# Copyright 2021 The Hush Developers
# Distributed under the GPLv3 software license, see the accompanying
# file LICENSE or https://www.gnu.org/licenses/gpl-3.0.en.html
# Purpose: Script to build the ARM or aarch64 architecture,
# which is what the PinePhone & newer Raspberry Pi boards are running

set -e

# Check if go is installed on system and exits if it is not
if ! [ -x "$(command -v go)" ]; then
  echo 'Error: go is not installed. Install go and try again.' >&2
  exit 1
fi

# Check for correct version of go installed on system and exits if it is not
GO_VERSION=`go version | { read _ _ v _; echo ${v#go}; }`
GO_VERSION_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
GO_VERSION_MINOR=$(echo $GO_VERSION | cut -d. -f2)
if [[ "$GO_VERSION_MINOR" -lt 17 ]]; then
  echo 'Error: go version 1.17 or greater is required to build lightwallted & that is not installed. Install the correct version and try again.' >&2
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
echo "You have the correct version of go installed, so starting to compile Hush lightwalletd for you..."
env GOOS=linux GOARCH=arm64 go build -o lightwalletd_aarch64 main.go
mv lightwalletd_aarch64 lightwalletd
echo ""
echo "Hush lightwalletd is now compiled for your ARM farm or Pine Rock64. Enjoy and reach out if you need support."
echo "For options, run ./lightwalletd --help"
