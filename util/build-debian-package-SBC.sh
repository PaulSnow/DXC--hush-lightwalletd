#!/usr/bin/env bash
# Copyright (c) 2021-2023 The Hush developers
# Distributed under the GPLv3 software license, see the accompanying
# file LICENSE or https://www.gnu.org/licenses/gpl-3.0.en.html
#
# Remix for SBC (Single Board Computer) like PineBook, Rock64, Raspberry Pi, etc.
## Usage: ./util/build-debian-package-SBC.sh

echo "Let's see who read the README.md or not..."
echo ""

# Check if lightwalletd is already built on system and exit if it is not
if ! [ -x "$(command -v ./lightwalletd)" ]; then
  echo 'Error: lightwalletd is not compiled yet. Run "make build" and try again.' >&2
  echo ""
  exit 1
fi
# Check if lintian is installed and exit if it is not
if ! [ -x "$(command -v lintian)" ]; then
  echo 'Error: lintian is not installed yet. Consult your Linux version package manager...' >&2
  echo 'On Debian/Ubuntu, try "sudo apt install lintian"'
  echo ""
  exit 1
fi
# Check if fakeroot is installed and exit if it is not
if ! [ -x "$(command -v fakeroot)" ]; then
  echo 'Error: fakeroot is not installed yet. Consult your Linux version package manager...' >&2
  echo 'On Debian/Ubuntu, try "sudo apt install fakeroot"'
  echo ""
  exit 1
fi

# TODO - check that the lightwalletd binary is not x86 and is actually aarch64

echo "Let There Be Hush Lightwalletd Debian Packages for ARM!!!"
echo ""
echo "((_,...,_))"
echo "   |o o|"
echo "   \   /"
echo "    ^_^   cp97"
echo ""

set -e
set -x

BUILD_PATH="/tmp/lightwalletd-debian-$$"
PACKAGE_NAME="lightwalletd"
SRC_PATH=`pwd`
SRC_DEB=$SRC_PATH/contrib/debian
SRC_DOC=$SRC_PATH/doc
ARCH="aarch64"

umask 022

if [ ! -d $BUILD_PATH ]; then
    mkdir $BUILD_PATH
fi

PACKAGE_VERSION=0.1.3
DEBVERSION=$(echo $PACKAGE_VERSION)
BUILD_DIR="$BUILD_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH"

if [ -d $BUILD_DIR ]; then
    rm -R $BUILD_DIR
fi

DEB_BIN=$BUILD_DIR/usr/bin
DEB_CMP=$BUILD_DIR/usr/share/bash-completion/completions
DEB_DOC=$BUILD_DIR/usr/share/doc/$PACKAGE_NAME
DEB_MAN=$BUILD_DIR/usr/share/man/man1
DEB_SHR=$BUILD_DIR/usr/share/hush
mkdir -p $BUILD_DIR/DEBIAN $DEB_CMP $DEB_BIN $DEB_DOC $DEB_MAN $DEB_SHR
chmod 0755 -R $BUILD_DIR/*

# Package maintainer scripts (currently empty)
#cp $SRC_DEB/postinst $BUILD_DIR/DEBIAN
#cp $SRC_DEB/postrm $BUILD_DIR/DEBIAN
#cp $SRC_DEB/preinst $BUILD_DIR/DEBIAN
#cp $SRC_DEB/prerm $BUILD_DIR/DEBIAN

# Copy binary
cp $SRC_PATH/lightwalletd $DEB_BIN/lightwalletd
strip $DEB_BIN/lightwalletd
cp $SRC_DEB/changelog $DEB_DOC
cp $SRC_DEB/copyright $DEB_DOC
cp -r $SRC_DEB/examples $DEB_DOC
# Copy manpage
cp $SRC_DOC/man/lightwalletd.1 $DEB_MAN/lightwalletd.1

# Gzip files
gzip --best -n $DEB_MAN/lightwalletd.1

cd $SRC_PATH/contrib

# Create the control file
dpkg-shlibdeps $DEB_BIN/lightwalletd
dpkg-gencontrol -P$BUILD_DIR -v$DEBVERSION

# Create the Debian package
fakeroot dpkg-deb --build $BUILD_DIR
cp $BUILD_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH.deb $SRC_PATH
shasum -a 256 $SRC_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH.deb
# Analyze with Lintian, reporting bugs and policy violations
lintian -i $SRC_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH.deb
exit 0
