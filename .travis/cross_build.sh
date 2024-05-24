#!/usr/bin/env bash

# Functions.
function info() {
    echo -e "\033[00;34mi\033[0m $1"
}

function fail() {
    echo -e "\033[00;31m!\033[0m $1"
    exit 1
}

function build() {
    goos=$1
    goarch=$2

    info "Building for $goos $goarch..."
    GOOS=$goos GOARCH=$goarch go build -o $goos_$goarch main.go
    if [[ $? -ne 0 ]]; then
        fail "Could not build for $goos $goarch. Aborting."
    fi
}

# Create build directory.
mkdir -p bin
cd bin

# Create go.mod
cat <<EOF > go.mod
module cross_build
require github.com/carmel/unipdf v3.0.0
EOF

echo "replace github.com/carmel/unipdf => $TRAVIS_BUILD_DIR" >> go.mod

# Create Go file.
cat <<EOF > main.go
package main

import (
	_ "github.com/carmel/unipdf/annotator"
	_ "github.com/carmel/unipdf/common"
	_ "github.com/carmel/unipdf/common/license"
	_ "github.com/carmel/unipdf/contentstream"
	_ "github.com/carmel/unipdf/contentstream/draw"
	_ "github.com/carmel/unipdf/core"
	_ "github.com/carmel/unipdf/core/security"
	_ "github.com/carmel/unipdf/core/security/crypt"
	_ "github.com/carmel/unipdf/creator"
	_ "github.com/carmel/unipdf/extractor"
	_ "github.com/carmel/unipdf/fdf"
	_ "github.com/carmel/unipdf/fjson"
	_ "github.com/carmel/unipdf/model"
	_ "github.com/carmel/unipdf/model/optimize"
	_ "github.com/carmel/unipdf/model/sighandler"
	_ "github.com/carmel/unipdf/ps"
	_ "github.com/carmel/unipdf/render"
)

func main() {}
EOF

# Build file.
for os in "linux" "darwin" "windows"; do
    for arch in "386" "amd64"; do
        build $os $arch
    done
done
