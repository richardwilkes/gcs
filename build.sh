#! /usr/bin/env bash
set -eo pipefail

trap 'echo -e "\033[33;5mBuild failed on build.sh:$LINENO\033[0m"' ERR

# Process args
RELEASE="0.0"
for arg in "$@"; do
  case "$arg" in
  --all | -a)
    LINT=1
    TEST=1
    RACE=-race
    ;;
  --dist|-d)
    EXTRA_BUILD_FLAGS="-a -trimpath"
    RELEASE="5.0.0"
    ;;
  --lint | -l) LINT=1 ;;
  --race | -r)
    TEST=1
    RACE=-race
    ;;
  --test | -t) TEST=1 ;;
  --help | -h)
    echo "$0 [options]"
    echo "  -a, --all  Equivalent to --lint --race"
    echo "  -d, --dist Build for distribution"
    echo "  -l, --lint Run the linters"
    echo "  -r, --race Run the tests with race-checking enabled"
    echo "  -t, --test Run the tests"
    echo "  -h, --help This help text"
    exit 0
    ;;
  *)
    echo "Invalid argument: $arg"
    exit 1
    ;;
  esac
done

echo -e "\033[33mBuilding...\033[0m"
LDFLAGS_ALL="-X github.com/richardwilkes/toolbox/cmdline.AppVersion=$RELEASE"
STD_FLAGS="-v -buildvcs $EXTRA_BUILD_FLAGS"

case $(uname -s) in
Darwin*)
  if [ "$(uname -p)" == "arm" ]; then
    export MACOSX_DEPLOYMENT_TARGET=11
  else
    export MACOSX_DEPLOYMENT_TARGET=10.14
  fi
esac

# Generate the source
go generate ./gen/enumgen.go

# Build our code
case $(uname -s) in
Darwin*)
  /bin/rm -rf GCS.app
  CONTENTS="GCS.app/Contents"
  mkdir -p "$CONTENTS/MacOS"
  mkdir -p "$CONTENTS/Resources"
  cp bundle/*.icns "$CONTENTS/Resources/"
  go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL" -o "$CONTENTS/MacOS/" .
  sed -e "s/SHORT_APP_VERSION/$($CONTENTS/MacOS/gcs -v | tr -d "\n")/" \
    -e "s/LONG_APP_VERSION/$($CONTENTS/MacOS/gcs -V | tr -d "\n")/" \
    -e "s/COPYRIGHT_YEARS/$($CONTENTS/MacOS/gcs --copyright-date | tr -d "\n")/" \
    bundle/Info.plist >"$CONTENTS/Info.plist"
  ;;
Linux*)
  go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL" .
  ;;
MINGW*)
  go install github.com/tc-hib/go-winres@v0.3.0
  if [ -e "$GOPATH/bin/go-winres.exe" ]; then
    GOWINRES="$GOPATH/bin/go-winres"
  else
    GOWINRES="$HOME/go/bin/go-winres"
  fi
  "$GOWINRES" make --arch amd64
  go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL -H windowsgui" .
  ;;
*)
  echo "Unsupported OS"
  false
  ;;
esac

# Run the tests
if [ "$TEST"x == "1x" ]; then
  if [ -n "$RACE" ]; then
    echo -e "\033[32mTesting with -race enabled...\033[0m"
  else
    echo -e "\033[32mTesting...\033[0m"
  fi
  go test $RACE ./...
fi

# Run the linters
if [ "$LINT"x == "1x" ]; then
  GOLANGCI_LINT_VERSION=1.48.0
  TOOLS_DIR=$PWD/tools
  mkdir -p "$TOOLS_DIR"
  if [ ! -e "$TOOLS_DIR/golangci-lint" ] || [ "$("$TOOLS_DIR/golangci-lint" version 2>&1 | awk '{ print $4 }' || true)x" != "${GOLANGCI_LINT_VERSION}x" ]; then
    echo -e "\033[33mInstalling version $GOLANGCI_LINT_VERSION of golangci-lint into $TOOLS_DIR...\033[0m"
    curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$TOOLS_DIR" v$GOLANGCI_LINT_VERSION
  fi
  echo -e "\033[32mLinting...\033[0m"
  $TOOLS_DIR/golangci-lint run
fi
