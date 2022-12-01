#!/bin/bash
# Copyright 2020-2022 The Hush Developers
# Released under GPLv3

# Description: This script would be used with a NGINX reverse proxy
#	you can choose either IPv4 or IPv6

# using ipv4 localhost
#./lightwalletd -bind-addr localhost:9067 -conf-file ~/.hush/HUSH3/HUSH3.conf -no-tls

# using ipv6 localhost
./lightwalletd -bind-addr ip6-localhost:9067 -conf-file ~/.hush/HUSH3/HUSH3.conf -no-tls
