#!/bin/bash -e

mkdir -p release
for os in linux darwin; do
    for arch in amd64 386; do
        echo "building for ${os} ${arch}"
        CGO_ENABLED=0 GOARCH=${arch} GOOS=${os} go build -ldflags '-X github.com/tjamet/kubectl-switch/update.Version='$(git describe --tags --dirty) -o release/kubectl-switch-${os}-${arch} .
    done
    cp release/kubectl-switch-${os}-amd64 release/kubectl-switch-${os}-x86_64
done

for f in $(ls ./release/); do
    shasum -a 1 ./release/${f} | awk '{print $1}' > ./release/${f}.sha1sum
    shasum -a 256 ./release/${f} | awk '{print $1}' > ./release/${f}.sha256sum
done