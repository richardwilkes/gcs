#! /usr/bin/env bash
set -eo pipefail

trap 'echo -e "\033[33;5mBuild failed on build.sh:$LINENO\033[0m"' ERR

VERSION="4.3.136"
ARCHIVE="pdfjs-${VERSION}-dist.zip"

/bin/rm -rf build web LICENSE pdf.js-version-*

curl -s -L https://github.com/mozilla/pdf.js/releases/download/v${VERSION}/${ARCHIVE} -o ${ARCHIVE}
unzip -qq ${ARCHIVE}
/bin/rm -f build/*.map web/*.map web/*.pdf web/debugger.* ${ARCHIVE}

touch pdf.js-version-${VERSION}