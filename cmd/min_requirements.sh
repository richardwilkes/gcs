#!/usr/bin/env bash
#
# Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, version 2.0. If a copy of the MPL was not distributed with
# this file, You can obtain one at http://mozilla.org/MPL/2.0/.
#
# This Source Code Form is "Incompatible With Secondary Licenses", as
# defined by the Mozilla Public License, version 2.0.
#
# Detect the minimum OS requirements baked into built GCS binaries.
#
# Usage:
#   ./min_requirements.sh <dir>            Analyze every binary found under <dir>
#   ./min_requirements.sh --run <id>       Download the artifacts from a GitHub
#                                          Actions run (via gh) and analyze them
#   ./min_requirements.sh --run latest     Use the most recent run on this branch
#   ./min_requirements.sh --release        Analyze the latest published GitHub
#                                          release's binaries
#   ./min_requirements.sh --release <tag>  Analyze a specific release (e.g. v5.42.0)
#
# What it reads from each binary:
#   Linux (ELF)    glibc:  highest GLIBC_x.y symbol version referenced
#                  kernel: minimum from the .note.ABI-tag section
#   macOS (Mach-O) minos:  LC_BUILD_VERSION / LC_VERSION_MIN_MACOSX
#   Windows (PE)   subsystem version (see the caveat printed below)

set -euo pipefail

err() {
	echo "error: $*" >&2
	exit 1
}

need() { command -v "$1" >/dev/null 2>&1 || err "required tool not found: $1"; }

# Map a required glibc x.y version to the oldest Ubuntu / Debian / Fedora
# release that satisfies it (i.e. ships that glibc version or newer).
glibc_distro() {
	case "$1" in
	2.41) echo "Ubuntu 25.04 / Debian 13 / Fedora 42 (or newer)" ;;
	2.40) echo "Ubuntu 24.10 / Debian 13 / Fedora 41 (or newer)" ;;
	2.39) echo "Ubuntu 24.04 / Debian 13 / Fedora 40 (or newer)" ;;
	2.38) echo "Ubuntu 23.10 / Debian 13 / Fedora 39 (or newer)" ;;
	2.37) echo "Ubuntu 23.04 / Debian 13 / Fedora 38 (or newer)" ;;
	2.36) echo "Ubuntu 22.10 / Debian 12 / Fedora 37 (or newer)" ;;
	2.35) echo "Ubuntu 22.04 / Debian 12 / Fedora 36 (or newer)" ;;
	2.34) echo "Ubuntu 21.10 / Debian 12 / Fedora 35 (or newer)" ;;
	2.33) echo "Ubuntu 21.04 / Debian 12 / Fedora 34 (or newer)" ;;
	2.32) echo "Ubuntu 20.10 / Debian 12 / Fedora 33 (or newer)" ;;
	2.31) echo "Ubuntu 20.04 / Debian 11 / Fedora 32 (or newer)" ;;
	2.30) echo "Ubuntu 19.10 / Debian 11 / Fedora 31 (or newer)" ;;
	2.29) echo "Ubuntu 19.04 / Debian 11 / Fedora 30 (or newer)" ;;
	2.28) echo "Ubuntu 18.10 / Debian 10 / Fedora 29 (or newer)" ;;
	2.27) echo "Ubuntu 18.04 / Debian 10 / Fedora 28 (or newer)" ;;
	*) echo "(distro lookup not tabulated)" ;;
	esac
}

# Map a macOS minos version to its marketing name. Releases 10.x carry a minor
# (e.g. 10.15); 11 and later are identified by their major version alone (Apple
# jumped from 15 "Sequoia" to 26 "Tahoe").
macos_name() {
	case "$1" in
	10.13*) echo "High Sierra" ;;
	10.14*) echo "Mojave" ;;
	10.15*) echo "Catalina" ;;
	11.*) echo "Big Sur" ;;
	12.*) echo "Monterey" ;;
	13.*) echo "Ventura" ;;
	14.*) echo "Sonoma" ;;
	15.*) echo "Sequoia" ;;
	26.*) echo "Tahoe" ;;
	*) echo "(unrecognized)" ;;
	esac
}

# Map a Windows PE major.minor version to a marketing name.
windows_name() {
	case "$1" in
	# Windows 11 shares the NT 10.0 version and is distinguished only by build
	# number (>= 22000), so a PE subsystem version of 10.0 covers both 10 and
	# 11 today.
	10.0) echo "Windows 10 / 11 / Server 2016+" ;;
	6.3) echo "Windows 8.1 / Server 2012 R2" ;;
	6.2) echo "Windows 8 / Server 2012" ;;
	6.1) echo "Windows 7 / Server 2008 R2" ;;
	6.0) echo "Windows Vista / Server 2008" ;;
	*) echo "(unrecognized)" ;;
	esac
}

# Shared column layout for every platform's row:
#   platform   arch   primary-requirement   secondary-requirement   [resolved name]
# Only Linux uses the secondary (kernel) column; others leave it blank so the
# resolved-name column lines up across all rows.
row_fmt='%-9s %-8s %-16s %-18s %s\n'

analyze_elf() {
	local f=$1 arch glibc kernel
	arch=$(file -b "$f" | grep -oE 'x86-64|aarch64|ARM aarch64' | head -1)
	[ "$arch" = "ARM aarch64" ] && arch=aarch64

	# Highest referenced GLIBC_x.y symbol version == minimum glibc required.
	glibc=$(objdump -T "$f" 2>/dev/null |
		grep -oE 'GLIBC_[0-9]+\.[0-9]+' | sed 's/GLIBC_//' |
		sort -t. -k1,1n -k2,2n | tail -1 || true)
	if [ -z "$glibc" ]; then
		glibc="(none - statically linked?)"
	fi

	# Minimum kernel from the ELF ABI-tag note (file prints "for GNU/Linux X.Y.Z").
	kernel=$(file -b "$f" | grep -oE 'GNU/Linux [0-9.]+' | grep -oE '[0-9.]+' | head -1)
	[ -z "$kernel" ] && kernel="(unspecified)"

	printf "$row_fmt" \
		"Linux" "$arch" "glibc >= $glibc" "kernel >= $kernel" "[$(glibc_distro "$glibc")]"
}

analyze_macho() {
	local f=$1 arch minos
	arch=$(file -b "$f" | grep -oE 'x86_64|arm64' | head -1)
	# vtool reports LC_BUILD_VERSION (minos) or the older LC_VERSION_MIN_MACOSX (version).
	minos=$(vtool -show-build "$f" 2>/dev/null | awk '/minos|version/ {print $2; exit}')
	[ -z "$minos" ] && minos="(unspecified)"
	printf "$row_fmt" "macOS" "$arch" "macOS >= $minos" "" "[$(macos_name "$minos")]"
}

analyze_pe() {
	local f=$1 arch major minor ver
	arch=$(file -b "$f" | grep -oiE 'x86-64|Aarch64' | head -1)
	[ "$arch" = "Aarch64" ] && arch=aarch64
	major=$(objdump -p "$f" 2>/dev/null | awk '/MajorSubsystemVersion/ {print $2; exit}')
	minor=$(objdump -p "$f" 2>/dev/null | awk '/MinorSubsystemVersion/ {print $2; exit}')
	ver="${major:-?}.${minor:-?}"
	printf "$row_fmt" "Windows" "$arch" "subsystem $ver" "" "[$(windows_name "$ver")]"
}

analyze() {
	local f=$1 kind
	kind=$(file -b "$f")
	# Only analyze the main application executables; skip shared libraries
	# (e.g. an embedded Skia DLL/dylib) so they don't add misleading rows.
	case "$kind" in
	*Mach-O*executable*) analyze_macho "$f" ;;
	*PE32*DLL*) ;; # skip Windows DLLs
	*PE32*) analyze_pe "$f" ;;
	*ELF*shared*) ;; # skip ELF shared objects (.so)
	*ELF*) analyze_elf "$f" ;;
	*) ;; # not an executable we care about
	esac
}

# Expand any release archives found in the directory in place, so the analyzer
# can walk the binaries inside them. Handles .tgz/.tar.gz, .zip, and macOS .dmg.
extract_archives() {
	local dir=$1 archive mnt dest
	while IFS= read -r -d '' archive; do
		case "$archive" in
		*.tgz | *.tar.gz)
			dest="$archive.extracted"
			mkdir -p "$dest"
			tar -xzf "$archive" -C "$dest" 2>/dev/null || true
			;;
		*.zip)
			dest="$archive.extracted"
			mkdir -p "$dest"
			unzip -o -q "$archive" -d "$dest" 2>/dev/null || true
			;;
		*.dmg)
			command -v hdiutil >/dev/null 2>&1 || {
				echo "warning: hdiutil not available; skipping $archive" >&2
				continue
			}
			dest="$archive.extracted"
			mkdir -p "$dest"
			# Mount read-only, copy the .app out, then detach. The mount point
			# is the trailing /Volumes/... path on the last line of the output.
			mnt=$(hdiutil attach -nobrowse -readonly "$archive" 2>/dev/null |
				sed -nE 's#.*(/Volumes/.*)$#\1#p' | tail -1)
			if [ -n "$mnt" ]; then
				cp -R "$mnt"/*.app "$dest"/ 2>/dev/null || true
				hdiutil detach -quiet "$mnt" 2>/dev/null || true
			fi
			;;
		esac
	done < <(find "$dir" -type f \( -name '*.tgz' -o -name '*.tar.gz' -o -name '*.zip' -o -name '*.dmg' \) -print0)
}

need file
need objdump

dir=""

case "${1:-}" in
--run)
	need gh
	run_id=${2:-}
	[ -n "$run_id" ] || err "usage: $0 --run <id|latest>"
	if [ "$run_id" = "latest" ]; then
		run_id=$(gh run list --limit 1 --json databaseId -q '.[0].databaseId')
		[ -n "$run_id" ] || err "no runs found"
	fi
	dir=$(mktemp -d)
	echo "Downloading artifacts from run $run_id into $dir ..." >&2
	gh run download "$run_id" --dir "$dir" >&2
	;;
--release)
	need gh
	tag=${2:-}
	dir=$(mktemp -d)
	if [ -n "$tag" ]; then
		echo "Downloading release $tag into $dir ..." >&2
		gh release download "$tag" --dir "$dir" --pattern '*' >&2
	else
		echo "Downloading latest release into $dir ..." >&2
		gh release download --dir "$dir" --pattern '*' >&2
	fi
	extract_archives "$dir"
	;;
"" | -h | --help)
	err "usage: $0 <dir> | --run <id|latest> | --release [tag]"
	;;
*)
	dir=$1
	[ -d "$dir" ] || err "not a directory: $dir"
	;;
esac

echo
echo "Minimum requirements detected from binaries in $dir:"
echo "-----------------------------------------------------------------------------"
# Collect every row, then sort so platforms group together for easy reading.
results=""
while IFS= read -r -d '' f; do
	row=$(analyze "$f")
	[ -n "$row" ] && results+="$row"$'\n'
done < <(find "$dir" -type f -print0)
printf '%s' "$results" | sort -u

if printf '%s' "$results" | grep -q '^Windows'; then
	echo
	echo "NOTE: The Windows PE subsystem version is hard-coded by the Go linker (it has"
	echo "      long reported 6.1 / Windows 7) and does NOT reflect the true runtime"
	echo "      minimum. Modern Go requires Windows 10 / Server 2016 or newer, so treat"
	echo "      Windows 10 as the real floor regardless of the value shown above."
fi
