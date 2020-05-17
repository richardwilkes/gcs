#!/bin/bash

set -eo pipefail

BOOTSTRAP_DIR=out/bootstrap

/bin/rm -rf "$BOOTSTRAP_DIR"
mkdir -p "$BOOTSTRAP_DIR"
javac -d "$BOOTSTRAP_DIR" -encoding UTF8 ./bundler/bundler/Bundler.java
java -cp "$BOOTSTRAP_DIR" bundler.Bundler "$@"
