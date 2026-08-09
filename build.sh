#! /usr/bin/env bash
set -eo pipefail

trap 'echo -e "\033[33;5mBuild failed on build.sh:$LINENO\033[0m"' ERR

export GOEXPERIMENT=jsonv2

# Process args
RELEASE="0.0"
for arg in "$@"; do
	case "$arg" in
	--all | -a)
		BUILD_GO=1
		BUILD_GEN=1
		FMT=1
		LINT=1
		TEST=1
		RACE=-race
		PACKAGER=1
		SOMETHING=1
		;;
	--go | -g)
		BUILD_GO=1
		PACKAGER=1
		SOMETHING=1
		;;
	--gen | -G)
		BUILD_GEN=1
		SOMETHING=1
		;;
	--genpkg | -p)
		RUN_GENPKG=1
		SOMETHING=1
		;;
	--fmt | -f)
		FMT=1
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
		if [ -z "$GCS_RELEASE" ]; then
			echo "GCS_RELEASE must be set" >&2
			exit 1
		fi
		EXTRA_LD_FLAGS="-s -w"
		EXTRA_BUILD_FLAGS="-a -trimpath"
		RELEASE="$GCS_RELEASE"
		PACKAGER=1
		DIST=--dist
		BUILD_GO=1
		BUILD_GEN=1
		SOMETHING=1
		;;
	--help | -h)
		echo "$0 [options]"
		echo "  -a, --all    Equivalent to --gen --go --fmt --lint --race"
		echo "  -d, --dist   Create distribution"
		echo "  -f, --fmt    Verify the source formatting (gofumpt and goimports)"
		echo "  -g, --go     Build the Go code"
		echo "  -G, --gen    Generate the source"
		echo "  -p, --genpkg Generate the icons and packaging.yml file"
		echo "  -i, --i18n   Extract the localization template"
		echo "  -l, --lint   Run the linters"
		echo "  -r, --race   Run the tests, race-checking those that exercise concurrency"
		echo "  -t, --test   Run the tests"
		echo "  -h, --help   This help text"
		exit 0
		;;
	*)
		echo "Invalid argument: $arg"
		exit 1
		;;
	esac
done

if [ "$RUN_GENPKG"x == "1x" ]; then
	go run ./cmd/genpkg/main.go
fi

if [ "$SOMETHING"x != "1x" ]; then
	BUILD_GEN=1
	BUILD_GO=1
fi

case $(uname -s) in
Darwin*)
	export MACOSX_DEPLOYMENT_TARGET=11
	;;
MINGW*|MSYS*)
	WINDOWS=1
	EXTRA_LD_FLAGS="$EXTRA_LD_FLAGS -H windowsgui"
	;;
esac

LDFLAGS_ALL="-X github.com/richardwilkes/toolbox/v2/xos.AppVersion=$RELEASE $EXTRA_LD_FLAGS"
STD_FLAGS="-v -buildvcs=true $EXTRA_BUILD_FLAGS"

# Generate the source
if [ "$BUILD_GEN"x == "1x" ]; then
	echo -e "\033[33mGenerating...\033[0m"
	go generate ./cmd/enumgen/main.go
fi

# Generate the translation file
if [ "$I18N"x == "1x" ]; then
	# Ensure all dependencies are present in the module cache; otherwise the `go list -f "{{.Dir}}"` lookups below
	# resolve to empty strings for any module that hasn't been downloaded yet, silently excluding it from the scan.
	go mod download
	i18n $(go list -f "{{.Dir}}" -m github.com/richardwilkes/canvas) \
		$(go list -f "{{.Dir}}" -m github.com/richardwilkes/pdfview) \
		$(go list -f "{{.Dir}}" -m github.com/richardwilkes/rpgtools) \
		$(go list -f "{{.Dir}}" -m github.com/richardwilkes/toolbox/v2) \
		$(go list -f "{{.Dir}}" -m github.com/richardwilkes/unison) \
		.
fi

# Build our Go code
if [ "$BUILD_GO"x == "1x" ]; then
	# On Windows the app icon and the version info shown in Explorer's Properties dialog are embedded via a .syso
	# resource object that the Go linker only links in if it is already present when `go build` runs. That object is
	# produced by the packager, so on Windows it must be generated *before* the build; the packaging step below re-emits
	# the same file, but by then the build has already consumed it. Without this, packaged Windows binaries end up as
	# generic, info-less executables.
	if [ "$PACKAGER"x == "1x" ] && [ "$WINDOWS"x == "1x" ]; then
		echo -e "\033[33mGenerating Windows resources...\033[0m"
		go run ./cmd/pack/main.go --release "$RELEASE"
	fi
	echo -e "\033[33mBuilding the Go code...\033[0m"
	go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL" .
fi

# Both the formatting check and the linters come out of golangci-lint, so whichever of them runs first installs it.
ensure_golangci_lint() {
	if [ -n "$GOLANGCI_LINT" ]; then
		return
	fi
	GOLANGCI_LINT_VERSION=$(curl --head -s https://github.com/golangci/golangci-lint/releases/latest | grep -i location: | sed 's/^.*v//' | tr -d '\r\n')
	TOOLS_DIR=$(go env GOPATH)/bin
	if [ ! -e "$TOOLS_DIR/golangci-lint" ] || [ "$("$TOOLS_DIR/golangci-lint" version 2>&1 | awk '{ print $4 }' || true)x" != "${GOLANGCI_LINT_VERSION}x" ]; then
		echo -e "\033[33mInstalling version $GOLANGCI_LINT_VERSION of golangci-lint into $TOOLS_DIR...\033[0m"
		mkdir -p "$TOOLS_DIR"
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$TOOLS_DIR" v$GOLANGCI_LINT_VERSION
	fi
	GOLANGCI_LINT="$TOOLS_DIR/golangci-lint"
}

# Verify formatting
#
# This needs its own pass because `golangci-lint run` ignores the `formatters` section of .golangci.yml entirely -- only
# `golangci-lint fmt` consults it -- so without this the gofumpt and goimports settings declared there enforce nothing.
if [ "$FMT"x == "1x" ]; then
	ensure_golangci_lint
	echo -e "\033[33mChecking the formatting of the Go code...\033[0m"
	if ! "$GOLANGCI_LINT" fmt --diff; then
		echo -e "\033[31mRun 'golangci-lint fmt' to apply the formatting shown above.\033[0m" >&2
		exit 1
	fi
	echo "0 issues."
fi

# Lint the Go code
if [ "$LINT"x == "1x" ]; then
	ensure_golangci_lint
	echo -e "\033[33mLinting the Go code...\033[0m"
	"$GOLANGCI_LINT" run
fi

# Run the tests
#
# Race instrumentation slows the test packages down 3-10x, yet the detector can only ever report an access made while
# two or more goroutines are live, which almost none of the tests produce. So the full suite runs uninstrumented, and
# the race pass is limited to the TestRace wrappers (see model/gurps/race_coverage_test.go), which re-run just the
# tests that actually put multiple goroutines over shared state.
if [ "$TEST"x == "1x" ]; then
	echo -e "\033[33mTesting...\033[0m"
	go test ./... | grep -v "no test files"
	if [ -n "$RACE" ]; then
		echo -e "\033[33mRace-checking the concurrency tests...\033[0m"
		go test -race -run '^TestRace$' ./... | grep -Ev "no test files|no tests to run"
	fi
fi

# Package
if [ "$PACKAGER"x == "1x" ]; then
	go run ./cmd/pack/main.go --release $RELEASE $DIST
fi
