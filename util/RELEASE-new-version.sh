#!/usr/bin/env bash
# Copyright 2021-2023 Duke Leto and The Hush Developers
# Distributed under the GPLv3 software license, see the accompanying
# file LICENSE or https://www.gnu.org/licenses/gpl-3.0.en.html
# Purpose: To process a release for a new version of lightwalletd

# *** Run with ***
#     ./util/RELEASE-new-version.sh 0.1.3
# *** the above 0.1.3 is the version number you're building ***

# *** Will ONLY work on a Debian system to create these Debian packages

VERSION=$1
readonly VERSION
sed -i "s/0.1.3/$VERSION/g" util/build-debian-package.sh
sed -i "s/0.1.3/$VERSION/g" util/build-debian-package-SBC.sh
sed -i "s/0.1.3/$VERSION/g" common/common.go
sed -i "s/0.1.3/$VERSION/g" doc/man/lightwalletd.1

set -e

echo ""
echo "Let's build new release $VERSION of lightwalletd..."
echo ""
echo ""
echo "Building x86 binary"
make build
echo "Building x86 Debian package"
./util/build-debian-package.sh
echo "lightwalletd $VERSION x86 is done... now to build ARM..."

# TO-DO: automate ARM deb pkg build
#echo ""
#echo "Building ARM binary"
#make build-arm
#./util/build-debian-package-SBC.sh
#rm lightwalletd
#
echo "ALL DONE FOR lightwalletd $VERSION release"
