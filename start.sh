#!/bin/bash
# Copyright 2020-2022 The Hush Developers
# Released under GPLv3

# Description: This script would be used with a NGINX reverse proxy
#	you can choose either IPv4 or IPv6

# using ipv4 localhost
# NOTE: --no-tls is secure only if we use nginx as reverse proxy on localhost
./lightwalletd --grpc-bind-addr localhost:9067 --hush-conf-path ~/.hush/HUSH3/HUSH3.conf --no-tls

# Use this instead to use ipv6 localhost
#./lightwalletd -bind-addr ip6-localhost:9067 -conf-file ~/.hush/HUSH3/HUSH3.conf -no-tls
