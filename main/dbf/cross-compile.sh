#/bin/bash

PATH_SCRIPT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PATH_BUILD="$PATH_SCRIPT/dbf_release"
mkdir -p $PATH_BUILD

# Linux
export GOOS="linux"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/linux_64/dbf"

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/linux_32/dbf"

# Windows
export GOOS="windows"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/win_64/dbf.exe"

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/win_32/dbf.exe"

# Darwin
export GOOS="darwin"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/mac_64/dbf"

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go build -o "$PATH_BUILD/mac_32/dbf"