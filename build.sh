#!/bin/sh
set -eo pipefail

if [[ -z "${VERSION}" ]]; then
  echo "env VERSION is empty"
  exit 1
fi

build(){
    os=$1
    arch=$2
    echo "Building for OS $os, arch $arch..."
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags "-s -w" -installsuffix cgo -o bin/pasta
    echo ${VERSION} > ./cmd/version.txt
    cd bin
    gzip pasta -c > pasta_${1}_${arch}_${VERSION}.gz
    rm pasta
    cd ..
}

rm -rf bin/*

build darwin amd64
build darwin arm64

build linux 386
build linux amd64
build linux arm
build linux arm64

build windows 386
build windows amd64

