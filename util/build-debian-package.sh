#!/usr/bin/env bash
# Copyright (c) 2021-2024 The Hush developers
# Distributed under the GPLv3 software license, see the accompanying
# file LICENSE or https://www.gnu.org/licenses/gpl-3.0.en.html
#
## USAGE
#	./util/build-debian-package.sh <architecture flag> <version-you-are-building>
#
## 		Example to build debian pkg for amd64 (most common) for version 0.2.0:
#			./util/build-debian-package.sh --amd64 0.2.0
#
## 		Example to build debian pkg for ARM (aarch64) for version 0.2.0:
#			./util/build-debian-package.sh --arm 0.2.0
#
## USAGE Requirements:
#	- Needs to be run in lightwalletd root directory 
#	- Needs to be run on a Debian system (NOT UBUNTU)

echo "Let's see who read the USAGE EXAMPLES in this script or not..."
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

## Command line options section
# Check if there are no CLI options entered and exit if so
if [ -z "$1" ] || [ -z "$2" ]; then
  echo 'YOU DOING IT WRONG...' >&2
  echo 'Read the Usage Examples at top of this script & TRY AGAIN' >&2
  exit 1
fi
# Architecture CLI option
if [ "$1" = "--amd64" -o "$1" = "--a64" ]; then
  ARCH="amd64"
elif [ "$1" = "--arm" -o "$1" = "--ARM" -o "$1" = "--aarch64" ]; then
  ARCH="aarch64"
fi
# Set Version-to-build from Second CLI option
PACKAGE_VERSION=$2

echo "Let There Be Hush Lightwalletd Debian Packages!"
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

umask 022

if [ ! -d $BUILD_PATH ]; then
    mkdir $BUILD_PATH
fi

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
# had to comment line below to move forward in this build script...
#dpkg-shlibdeps $DEB_BIN/lightwalletd
dpkg-gencontrol -P$BUILD_DIR -v$DEBVERSION

# Create the Debian package
fakeroot dpkg-deb --build $BUILD_DIR
cp $BUILD_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH.deb $SRC_PATH
shasum -a 256 $SRC_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH.deb
# Analyze with Lintian, reporting bugs and policy violations
lintian -i $SRC_PATH/$PACKAGE_NAME-$PACKAGE_VERSION-$ARCH.deb
exit 0
