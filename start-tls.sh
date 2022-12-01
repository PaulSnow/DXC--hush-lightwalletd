#!/bin/bash
# Copyright 2020-2022 The Hush Developers
# Released under GPLv3

# Description: This script would be used with a NGINX reverse proxy

./lightwalletd --grpc-bind-addr localhost:9067 --hush-conf-path ~/.hush/HUSH3/HUSH3.conf $@

