#!/bin/bash

set -eo pipefail

/bin/rm -rf GCS out GCS-*.AppImage
mkdir -p GCS
cp ../*.deb GCS/
pkg2appimage-1807-x86_64.AppImage gcs_recipe.yml
mv out/GCS-*.AppImage $(ls out/GCS-*.AppImage | sed -E -e 's@out/@@' -e 's/\.glibc[0-9]+\.[0-9]+//')
/bin/rm -rf GCS out
