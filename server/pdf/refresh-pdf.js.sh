#! /usr/bin/env bash
set -eo pipefail

trap 'echo -e "\033[33;5mBuild failed on build.sh:$LINENO\033[0m"' ERR

VERSION="4.3.136"

/bin/rm -rf build web LICENSE pdf.js-version-*

curl -s -L https://github.com/mozilla/pdf.js/releases/download/v${VERSION}/pdfjs-${VERSION}-dist.zip | tar xz
/bin/rm -f build/*.map web/*.map web/*.pdf web/debugger.*

touch pdf.js-version-${VERSION}