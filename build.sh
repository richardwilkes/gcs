#! /usr/bin/env bash
set -eo pipefail

trap 'echo -e "\033[33;5mBuild failed on build.sh:$LINENO\033[0m"' ERR

# Process args
RELEASE="0.0"
for arg in "$@"; do
  case "$arg" in
  --all | -a)
	BUILD_JS=1
    BUILD_GO=1
	BUILD_GEN=1
    LINT=1
    TEST=1
    RACE=-race
	SOMETHING=1
    ;;
  --go | -g)
    BUILD_GO=1
	SOMETHING=1
	;;
  --js | -j)
    BUILD_JS=1
	SOMETHING=1
    ;;
  --gen | -G)
    BUILD_GEN=1
	SOMETHING=1
	;;
  --lint | -l)
    LINT=1
	SOMETHING=1
	;;
  --test | -t)
    TEST=1
	SOMETHING=1
	;;
  --race | -r)
    TEST=1
    RACE=-race
	SOMETHING=1
    ;;
  --i18n | -i)
    I18N=1
	SOMETHING=1
	;;
  --dist | -d)
    EXTRA_BUILD_FLAGS="-a -trimpath"
    EXTRA_LD_FLAGS="-s -w"
    RELEASE="5.20.4"
    DIST=1
	SOMETHING=1
    case $(uname -s) in
    Darwin*)
      if [ "$(uname -p)" == "arm" ]; then
        export MACOSX_DEPLOYMENT_TARGET=11
      else
        export MACOSX_DEPLOYMENT_TARGET=10.15
      fi
    esac
    ;;
  --help | -h)
    echo "$0 [options]"
    echo "  -a, --all  Equivalent to --gen --js --go --lint --race"
    echo "  -d, --dist Create distribution"
	echo "  -g, --go   Build the Go code"
	echo "  -G, --gen  Generate the source"
    echo "  -i, --i18n Extract the localization template"
	echo "  -j, --js   Build the JavaScript code"
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

if [ "$SOMETHING"x != "1x" ]; then
  echo -e "\033[33;5mNo options specified!\033[0m"
  exit 1
fi

LDFLAGS_ALL="-X github.com/richardwilkes/toolbox/cmdline.AppVersion=$RELEASE $EXTRA_LD_FLAGS"
STD_FLAGS="-v -buildvcs=true $EXTRA_BUILD_FLAGS"

# Build our JavaScript code
if [ "$BUILD_JS"x == "1x" ]; then
  echo -e "\033[33mBuilding the JavaScript code...\033[0m"
  cd server/frontend
  npm install
  npm run build
  if [ "$LINT"x == "1x" ]; then
    echo -e "\033[33mLinting the JavaScript code...\033[0m"
    npm run lint
  fi
  cd ../..
fi

# Generate the source
if [ "$BUILD_GEN"x == "1x" ]; then
  echo -e "\033[33mGenerating...\033[0m"
  go generate ./gen/enumgen.go
fi

# Generate the translation file
if [ "$I18N"x == "1x" ]; then
  i18n $(go list -f "{{.Dir}}" -m github.com/richardwilkes/json) \
    $(go list -f "{{.Dir}}" -m github.com/richardwilkes/pdf) \
    $(go list -f "{{.Dir}}" -m github.com/richardwilkes/rpgtools) \
    $(go list -f "{{.Dir}}" -m github.com/richardwilkes/toolbox) \
    $(go list -f "{{.Dir}}" -m github.com/richardwilkes/unison) \
    .
fi

# Build our Go code
if [ "$BUILD_GO"x == "1x" ]; then
  echo -e "\033[33mBuilding the Go code...\033[0m"
  case $(uname -s) in
  Darwin*)
    go run $STD_FLAGS -ldflags all="$LDFLAGS_ALL" packaging/main.go
    go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL" -o "GCS.app/Contents/MacOS/" .
    touch GCS.app
    ;;
  Linux*)
    go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL" .
    ;;
  MINGW*)
    go run $STD_FLAGS -ldflags all="$LDFLAGS_ALL" packaging/main.go
    go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL -H windowsgui" .
    ;;
  *)
    echo "Unsupported OS"
    false
    ;;
  esac
  if [ "$LINT"x == "1x" ]; then
    GOLANGCI_LINT_VERSION=$(curl --head -s https://github.com/golangci/golangci-lint/releases/latest | grep location: | sed 's/^.*v//' | tr -d '\r\n' )
    TOOLS_DIR=$(go env GOPATH)/bin
    if [ ! -e "$TOOLS_DIR/golangci-lint" ] || [ "$("$TOOLS_DIR/golangci-lint" version 2>&1 | awk '{ print $4 }' || true)x" != "${GOLANGCI_LINT_VERSION}x" ]; then
      echo -e "\033[33mInstalling version $GOLANGCI_LINT_VERSION of golangci-lint into $TOOLS_DIR...\033[0m"
      mkdir -p "$TOOLS_DIR"
      curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$TOOLS_DIR" v$GOLANGCI_LINT_VERSION
    fi
    echo -e "\033[33mLinting the Go code...\033[0m"
    "$TOOLS_DIR/golangci-lint" run
  fi
fi

# Run the tests
if [ "$TEST"x == "1x" ]; then
  if [ -n "$RACE" ]; then
    echo -e "\033[33mTesting with -race enabled...\033[0m"
  else
    echo -e "\033[33mTesting...\033[0m"
  fi
  go test $RACE ./... | grep -v "no test files"
fi

# Package for distribution
if [ "$DIST"x == "1x" ]; then
  echo -e "\033[33mPackaging...\033[0m"
  case $(uname -s) in
  Darwin*)
    if [ "$(uname -p)" == "arm" ]; then
      HW=apple
    else
      HW=intel
    fi
    codesign -s "Richard Wilkes" -f -v --timestamp --options runtime GCS.app
    /bin/rm -rf tmp *.dmg
    mkdir tmp
    # Installation of https://github.com/create-dmg/create-dmg is required
    create-dmg --volname "GCS v$RELEASE" --icon-size 128 --window-size 448 280 --add-file GCS.app GCS.app 64 64 \
      --app-drop-link 256 64 --codesign "Richard Wilkes" --hdiutil-quiet --no-internet-enable --notarize gcs-notary \
      gcs-$RELEASE-macos-$HW.dmg tmp
    /bin/rm -rf tmp
    ;;
  Linux*)
    /bin/rm -f gcs-${RELEASE}-linux.tgz
    tar czf gcs-${RELEASE}-linux.tgz gcs
    ;;
  MINGW*)
    go run -ldflags all="$LDFLAGS_ALL" packaging/main.go -z
    ;;
  esac
fi
