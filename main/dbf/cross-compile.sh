#/bin/bash

# Set the GOPATH if it is not set or if is different when running as another user (sudo)
# export GOPATH="" 

# Linux
export GOOS="linux"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install
echo "Done"
echo

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install
echo "Done"
echo

# Windows
export GOOS="windows"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install
echo "Done"
echo

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install
echo "Done"
echo

# Darwin
export GOOS="darwin"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install
echo "Done"
echo

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install
echo "Done"
echo