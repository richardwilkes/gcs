#!/bin/bash

echo Building module gnu.trove...
cd ../gnu.trove
./build.sh

echo Building module org.apache.commons.logging...
cd ../org.apache.commons.logging
./build.sh

echo Building module org.apache.fontbox...
cd ../org.apache.fontbox
./build.sh

echo Building module org.apache.pdfbox...
cd ../org.apache.pdfbox
./build.sh

echo Building module com.lowagie.text...
cd ../com.lowagie.text
./build.sh

echo Building module com.trollworks.toolkit...
cd ../toolkit
./build.sh

echo Building module com.trollworks.gcs...
cd ../gcs
./build.sh

echo Packaging GURPS Character Sheet...

/bin/rm -rf .mac
mkdir -p .mac/package/macosx
java --module-path ../java_modules --module com.trollworks.gcs/com.trollworks.gcs.app.GCSImages -icns -dir .mac
java --module-path ../java_modules --module com.trollworks.gcs/com.trollworks.gcs.app.GCSInfoPlistCreator .mac/package/macosx

/bin/rm -rf pkg
mkdir pkg
javapackager -deploy -native image --module-path ../java_modules --module com.trollworks.gcs -nosign -outdir pkg -name gcs -BdropinResourcesRoot=.mac

mv .mac/*.icns pkg/gcs.app/Contents/Resources/
mv .mac/package/macosx/PkgInfo pkg/gcs.app/Contents/
mv pkg/gcs.app "pkg/GURPS Character Sheet.app"

/bin/cp -R ../gcs_library/Library pkg/

/bin/rm -rf .mac