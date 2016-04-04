#/bin/bash

# Set the GOPATH if it is not set or if is different when running as another user (sudo)
# export GOPATH="" 

PATH_SCRIPT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PATH_BUILD="$GOPATH/bin/dbf_release"
mkdir -p $PATH_BUILD

# Linux
export GOOS="linux"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/linux_64"

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/linux_32"

# Windows
export GOOS="windows"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/win_64"

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/win_32"

# Darwin
export GOOS="darwin"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/mac_64"

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/mac_32"