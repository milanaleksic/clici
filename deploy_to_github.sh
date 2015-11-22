#!/bin/bash

# Dependencies:
# 1. go get github.com/aktau/github-release
# 2. go get github.com/pwaller/goupx
# 3. curl http://upx.sourceforge.net/download/upx-3.91-amd64_linux.tar.bz2 | tar xjvf - && sudo mv upx-3.91-amd64_linux/upx /usr/local/bin/ && rm -rf upx-3.91-amd64_linux

APPNAME=jenkins_ping

if [ "$GITHUB_TOKEN" = "" ]
then
    echo "GITHUB_TOKEN is not set!"
    exit 1
fi

if [ "$1" = "" ]
then
    echo "Which tag? No argument given to the application"
    exit 1
fi

echo Creating and pushing tag
git tag $1
git push --tags

echo Creating release
github-release release -u milanaleksic -r $APPNAME --tag "$1" --name "v$1"

# $1 - os
# $2 - suffix for executables
# $3 - tag
ship() {
    export GOOS=$1
    echo Building $GOOS
    go build
    echo Packing $1
    if [ "$1" = "linux" ]
    then
        goupx $APPNAME$2
    else
        upx $APPNAME$2
    fi
    echo Sending $1 to Github
    github-release upload -u milanaleksic -r $APPNAME --tag $3 --name "$APPNAME-$3-$1-amd64$2" -f $APPNAME$2
    rm $APPNAME$2
}

# ship "windows" ".exe" $1
ship "linux" "" $1
