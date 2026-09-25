#! /usr/bin/env bash
set -eEo pipefail

# -E above makes this trap fire for failures inside functions and subshells too, not just at the top level.
trap 'echo -e "\033[33;5mBuild failed on build.sh:$LINENO\033[0m"' ERR

# Everything below works relative to the root of the source tree, which is wherever this script lives. The tree may be a
# clone or a worktree of any name, and the script may be run from any directory.
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd "$ROOT"

# The root must be where the GCS module itself is defined: go has to find the go.mod for github.com/richardwilkes/gcs/v5
# here, not in some directory above. Workspace mode is turned off for the check, so that the answer describes this
# directory alone rather than whatever go.work happens to be in effect.
MODULE_PATH=github.com/richardwilkes/gcs/v5
if ! FOUND_MODULE=$(GOWORK=off go list -m -f '{{.Path}} {{.Dir}}' 2>/dev/null) ||
	[ "${FOUND_MODULE%% *}" != "$MODULE_PATH" ] ||
	[ "$(cd "${FOUND_MODULE#* }" && pwd -P)" != "$(pwd -P)" ]; then
	echo "$ROOT is not the root of the $MODULE_PATH module (go reports: ${FOUND_MODULE:-no module})" >&2
	exit 1
fi

export GOEXPERIMENT=simd

RELEASE="0.0"
while [ $# -gt 0 ]; do
	arg="$1"
	shift
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
	--headlessapi | -H)
		# The headless debug API (see ux/headless_api.md) only exists behind the headlessapi build tag. Setting the tag
		# through GOFLAGS puts it on every go invocation that follows -- the build, the tests and the linters -- so the
		# files behind it are built, tested and linted alongside everything else.
		export GOFLAGS="$GOFLAGS -tags=headlessapi"
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
	--smoke | -s)
		SMOKE=1
		SOMETHING=1
		;;
	--i18n | -i)
		I18N=1
		SOMETHING=1
		;;
	--target | -T | --target=*)
		if [ "$arg" == "${arg#--target=}" ]; then
			if [ $# -eq 0 ]; then
				echo "$arg requires an os/arch value, e.g. windows/amd64" >&2
				exit 1
			fi
			TARGET="$1"
			shift
		else
			TARGET="${arg#--target=}"
		fi
		if [ "$TARGET" == "${TARGET#*/}" ]; then
			echo "Invalid target: $TARGET (expected os/arch, e.g. windows/amd64)" >&2
			exit 1
		fi
		export GOOS="${TARGET%%/*}"
		export GOARCH="${TARGET#*/}"
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
		echo "  -a, --all            Equivalent to --gen --go --fmt --lint --race"
		echo "  -d, --dist           Create distribution"
		echo "  -f, --fmt            Verify the source formatting (gofumpt)"
		echo "  -g, --go             Build the Go code"
		echo "  -G, --gen            Generate the source"
		echo "  -H, --headlessapi    Build the Go code with the headless debug API compiled in"
		echo "  -p, --genpkg         Generate the icons and packaging.yml file"
		echo "  -i, --i18n           Extract the localization template"
		echo "  -l, --lint           Run the linters"
		echo "  -r, --race           Run the tests, race-checking those that exercise concurrency"
		echo "  -s, --smoke          Run the headless smoke tests. With GCS_SMOKE_WARN_ONLY set, a failure is only a"
		echo "                       warning"
		echo "  -t, --test           Run the tests"
		echo "  -T, --target OS/ARCH Build the application for another platform, e.g. windows/amd64. GOOS and GOARCH"
		echo "                       set in the environment work too. Packaging is only possible for this machine's"
		echo "                       own platform, so it is skipped (and --dist refused) when the target differs"
		echo "  -h, --help           This help text"
		exit 0
		;;
	*)
		echo "Invalid argument: $arg"
		exit 1
		;;
	esac
done

# A go.work in a parent directory, such as one tying the main checkout to sibling modules, is also found from within any
# worktree kept inside the tree, yet doesn't list that worktree, which leaves go unable to build it. Unless GOWORK was
# set explicitly, module mode is used instead whenever the workspace in effect doesn't include this directory. A
# workspace is also how a locally modified dependency (such as unison) gets built in, though, so dropping one that
# supplies or replaces a module this build needs would quietly build against the wrong code. In that case the build
# stops instead; setting GOWORK=off says that building against go.mod alone is intended.
check_workspace_dropped() {
	local work_file="$1" used replaced required path dir version needed=()
	used=$(cd "$(dirname "$work_file")" && GOWORK="$work_file" go list -m -f '{{.Path}} {{.Dir}}' 2>/dev/null || true)
	replaced=$(go work edit -json "$work_file" 2>/dev/null | grep -A1 '"Old"' |
		sed -n 's/.*"Path": "\([^"]*\)".*/\1/p' || true)
	if [ -n "$used" ] || [ -n "$replaced" ]; then
		if ! required=$(go list -m -f '{{if not .Main}}{{.Path}} {{.Version}}{{end}}' all); then
			echo -e "\033[31mUnable to list this build's dependencies to compare against $work_file. Set GOWORK=off to" \
				"build against go.mod alone.\033[0m" >&2
			exit 1
		fi
	fi
	while read -r path dir; do
		if [ -n "$path" ] && [ "$path" != "$MODULE_PATH" ]; then
			version=$(awk -v p="$path" '$1 == p { print $2 }' <<<"$required")
			if [ -n "$version" ]; then
				needed+=("$path: would use $version from go.mod rather than $dir")
			fi
		fi
	done <<<"$used"
	while read -r path; do
		if [ -n "$path" ]; then
			version=$(awk -v p="$path" '$1 == p { print $2 }' <<<"$required")
			if [ -n "$version" ]; then
				needed+=("$path: would use $version from go.mod, as its replace directive would not apply")
			fi
		fi
	done <<<"$replaced"
	if [ ${#needed[@]} -ne 0 ]; then
		echo -e "\033[31m$work_file supplies modules this build needs:\033[0m" >&2
		printf '  %s\n' "${needed[@]}" >&2
		echo -e "\033[31mAdd $ROOT to that workspace (or give it a go.work of its own) to build against them, or set" \
			"GOWORK=off to build against go.mod alone.\033[0m" >&2
		exit 1
	fi
	echo -e "\033[33mIgnoring $work_file, which doesn't include $ROOT and supplies nothing this build needs.\033[0m" >&2
}
# Output is captured before being searched, here and below, rather than piped into "grep -q": grep stops reading at the
# first match, which can kill the writer with SIGPIPE, and pipefail then reports the whole pipeline as failing.
if [ -z "${GOWORK+set}" ]; then
	WORK_FILE=$(go env GOWORK)
	if [ -n "$WORK_FILE" ]; then
		WORK_MODULE_DIRS=$(go list -m -f '{{.Dir}}' 2>/dev/null || true)
		if ! grep -Fqx -e "$ROOT" -e "$(pwd -P)" <<<"$WORK_MODULE_DIRS"; then
			export GOWORK=off
			check_workspace_dropped "$WORK_FILE"
		fi
	fi
fi

# The build machine and the target are told apart here. Only the application itself is built for the target: the tools
# run along the way (the enum generator and the packager), the tests and the linters all run on this machine, so from
# here on GOOS and GOARCH name the build machine's platform, and the application's build names the target explicitly.
HOST_OS=$(go env GOHOSTOS)
HOST_ARCH=$(go env GOHOSTARCH)
TARGET_OS=$(go env GOOS)
TARGET_ARCH=$(go env GOARCH)
SUPPORTED_TARGETS=$(go tool dist list)
if ! grep -Fqx "$TARGET_OS/$TARGET_ARCH" <<<"$SUPPORTED_TARGETS"; then
	echo "Unsupported target: $TARGET_OS/$TARGET_ARCH" >&2
	exit 1
fi
export GOOS="$HOST_OS"
export GOARCH="$HOST_ARCH"
if [ "$TARGET_OS/$TARGET_ARCH" != "$HOST_OS/$HOST_ARCH" ]; then
	CROSS=1
	echo -e "\033[33mCross-compiling for $TARGET_OS/$TARGET_ARCH on $HOST_OS/$HOST_ARCH\033[0m"
fi

# The packager (from unison) is built for a single platform and packages for that platform alone: the one it runs on.
if [ "$DIST"x != "x" ] && [ "$CROSS"x == "1x" ]; then
	echo "A distribution can only be created for this machine's own platform ($HOST_OS/$HOST_ARCH)" >&2
	exit 1
fi

# The tools are built for this machine rather than run with "go run", so that they are built once and never pick up a
# target's settings. None of them needs cgo, so they are built without it.
BUILD_TOOLS_DIR=$(mktemp -d)
trap 'rm -rf "$BUILD_TOOLS_DIR"' EXIT
tool() {
	local name="$1"
	shift
	if [ ! -e "$BUILD_TOOLS_DIR/$name$(go env GOEXE)" ]; then
		CGO_ENABLED=0 go build -o "$BUILD_TOOLS_DIR/$name$(go env GOEXE)" "./cmd/$name"
	fi
	"$BUILD_TOOLS_DIR/$name$(go env GOEXE)" "$@"
}

if [ "$RUN_GENPKG"x == "1x" ]; then
	tool genpkg
fi

if [ "$SOMETHING"x != "1x" ]; then
	BUILD_GEN=1
	BUILD_GO=1
fi

# Go 1.27+ hard-stamps its own macOS support floor (13.0 as of 1.27) into the intermediate go.o it hands to Apple's
# linker, which makes ld warn "object file (go.o) was built for newer 'macOS' version" on every link that targets 11. Go
# still runs fine on macOS 11 -- the floor only tracks which builders the Go team keeps around -- so stamp go.o with our
# real target instead. GOFLAGS covers the go test and go run links made for this machine; the application's build below
# passes explicit -ldflags, which override those in GOFLAGS, so EXTRA_LD_FLAGS needs it as well when the target is macOS.
if [ "$HOST_OS" == "darwin" ]; then
	export MACOSX_DEPLOYMENT_TARGET=11
	export GOFLAGS="$GOFLAGS -ldflags=-macos=11.0"
fi
case "$TARGET_OS" in
darwin)
	EXTRA_LD_FLAGS="$EXTRA_LD_FLAGS -macos=11.0"
	;;
windows)
	EXTRA_LD_FLAGS="$EXTRA_LD_FLAGS -H windowsgui"
	;;
esac

LDFLAGS_ALL="-X github.com/richardwilkes/toolbox/v2/xos.AppVersion=$RELEASE $EXTRA_LD_FLAGS"
STD_FLAGS="-v -buildvcs=true $EXTRA_BUILD_FLAGS"

if [ "$BUILD_GEN"x == "1x" ]; then
	echo -e "\033[33mGenerating...\033[0m"
	tool enumgen -root "$ROOT"
fi

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

if [ "$BUILD_GO"x == "1x" ]; then
	# On Windows the app icon and the version info Explorer shows come from a .syso resource object that the Go linker
	# only links in if it is already present when `go build` runs. The packager produces it, so it has to be generated
	# before the build; the packaging step below re-emits it, but by then the build has already consumed it. Without
	# this, packaged Windows binaries end up as generic, info-less executables. The packager writes one for each
	# Windows architecture, so this works for any Windows target, but only when running on Windows.
	if [ "$PACKAGER"x == "1x" ] && [ "$TARGET_OS" == "windows" ]; then
		if [ "$HOST_OS" == "windows" ]; then
			echo -e "\033[33mGenerating Windows resources...\033[0m"
			tool pack --release "$RELEASE"
		else
			echo -e "\033[33mWindows resources can only be generated on Windows; the executable will lack its icon and version info\033[0m"
		fi
	fi
	echo -e "\033[33mBuilding the Go code...\033[0m"
	GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build $STD_FLAGS -ldflags all="$LDFLAGS_ALL" .
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

# `golangci-lint run` also enforces the `formatters` section of .golangci.yml, but only for the files it loads, and no
# lint pass loads the files behind the race/!race build tags. `golangci-lint fmt` ignores build constraints, so this
# pass is what covers every file. It is also the one that prints the actual diff rather than a single "not properly
# formatted" issue per file.
if [ "$FMT"x == "1x" ]; then
	ensure_golangci_lint
	echo -e "\033[33mChecking the formatting of the Go code...\033[0m"
	if ! "$GOLANGCI_LINT" fmt --diff; then
		echo -e "\033[31mRun 'golangci-lint fmt' to apply the formatting shown above.\033[0m" >&2
		exit 1
	fi
	echo "0 issues."
fi

if [ "$LINT"x == "1x" ]; then
	ensure_golangci_lint
	echo -e "\033[33mLinting the Go code...\033[0m"
	"$GOLANGCI_LINT" run
fi

# Race instrumentation slows the test packages down 3-10x, yet the detector can only ever report an access made while
# two or more goroutines are live, which almost none of the tests produce. So the full suite runs uninstrumented and the
# race pass is limited to the TestRace wrappers (see the race_coverage_test.go files), which re-run just the tests
# that actually put multiple goroutines over shared state.
if [ "$TEST"x == "1x" ]; then
	echo -e "\033[33mTesting...\033[0m"
	go test ./... | grep -v "no test files"
	if [ -n "$RACE" ]; then
		echo -e "\033[33mRace-checking the concurrency tests...\033[0m"
		go test -race -run '^TestRace$' ./... | grep -Ev "no test files|no tests to run"
	fi
fi

# The smoke tests start the whole application headless against a fixed set of fixtures and compare what it shows with
# golden files (see ux/smoke_harness_test.go). They are compiled only with the smoke build tag, so the run above leaves
# them out. CI sets GCS_SMOKE_WARN_ONLY, which reports a failure as a warning rather than failing the build.
if [ "$SMOKE"x == "1x" ]; then
	echo -e "\033[33mRunning the smoke tests...\033[0m"
	if ! go test -tags smoke -run '^TestSmoke' ./ux; then
		if [ -z "$GCS_SMOKE_WARN_ONLY" ]; then
			echo -e "\033[33;5mThe smoke tests failed\033[0m"
			exit 1
		fi
		echo -e "\033[33mThe smoke tests failed; continuing, since GCS_SMOKE_WARN_ONLY is set\033[0m"
		if [ -n "$GITHUB_ACTIONS" ]; then
			echo "::warning title=Smoke tests::The smoke tests failed. See the build log, and the smoke test artifact for what they saw."
		fi
	fi
fi

if [ "$PACKAGER"x == "1x" ]; then
	if [ "$CROSS"x == "1x" ]; then
		echo -e "\033[33mSkipping packaging, which is only possible for this machine's own platform ($HOST_OS/$HOST_ARCH)\033[0m"
	else
		tool pack --release "$RELEASE" $DIST
	fi
fi
