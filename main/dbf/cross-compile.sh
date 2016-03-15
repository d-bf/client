#/bin/bash

# Set the GOPATH if it is not set or if is different when running as another user (sudo)
# export GOPATH="" 

# Linux
export GOOS="linux"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install

# Windows
export GOOS="windows"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install

# Darwin
export GOOS="darwin"

export GOARCH="amd64"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install

export GOARCH="386"
echo "Compiling... (OS: $GOOS, ARCH: $GOARCH)"
go install