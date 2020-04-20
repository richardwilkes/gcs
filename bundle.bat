@echo off
setlocal enableextensions
set BOOTSTRAP_DIR=out\bootstrap

if exist %BOOTSTRAP_DIR% rmdir /s /q %BOOTSTRAP_DIR%
mkdir %BOOTSTRAP_DIR%
javac -d %BOOTSTRAP_DIR% -encoding UTF8 .\bundler\bundler\Bundler.java
java -cp %BOOTSTRAP_DIR% bundler.Bundler
