#!/bin/bash

set -eo pipefail

VERSION=2.11.338

cd com.trollworks.gcs/resources/pdfjs
/bin/rm -rf build web LICENSE
curl --location --silent --output pdfjs-$VERSION-dist.zip https://github.com/mozilla/pdf.js/releases/download/v$VERSION/pdfjs-$VERSION-dist.zip
unzip -q pdfjs-$VERSION-dist.zip
/bin/rm -f web/*.pdf web/*.map build/*.map pdfjs-$VERSION-dist.zip
