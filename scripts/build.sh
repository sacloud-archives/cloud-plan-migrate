#!/bin/bash

set -e
#set -x

mkdir -p bin/ 2>/dev/null

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        arch="$GOOS-$GOARCH"
        binary="cloud-plan-migrate"
        if [ "$GOOS" = "windows" ]; then
          binary="${binary}.exe"
        fi
        echo "Building $binary $arch"
        GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 \
            go build \
                -ldflags "$BUILD_LDFLAGS" \
                -o bin/$binary \
                main.go
        if [ -n "$ARCHIVE" ]; then
            (cd bin/; zip -r "cloud-plan-migrate_$arch" $binary)
            rm -f bin/$binary
        fi
    done
done
