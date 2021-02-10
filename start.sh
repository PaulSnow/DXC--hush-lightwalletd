#!/bin/bash
# Copyright 2020-2021 The Hush Developers
# Released under GPLv3

# using ipv4 localhost
#./go/bin/go run cmd/server/main.go -bind-addr localhost:9067 -conf-file ~/.komodo/HUSH3/HUSH3.conf -no-tls

# using ipv6 localhost
./go/bin/go run cmd/server/main.go -bind-addr ip6-localhost:9067 -conf-file ~/.komodo/HUSH3/HUSH3.conf -no-tls
