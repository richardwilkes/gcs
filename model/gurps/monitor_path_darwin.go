// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/richardwilkes/toolbox/v2/xio"
)

// darwinPathMax is PATH_MAX from <sys/syslimits.h>, the most F_GETPATH will write.
const darwinPathMax = 1024

// resolvePath returns the given path as the OS names it: symlinks resolved and every component in the case it has on
// disk. filepath.EvalSymlinks resolves the symlinks but keeps the case it was given, and on the case-insensitive
// filesystems macOS uses by default the two differ whenever the path was typed rather than picked. F_GETPATH on an open
// descriptor asks the kernel for the path it knows the file by, which is also the form FSEvents reports changes in.
// Anything that keeps that from working falls back to what EvalSymlinks can do.
func resolvePath(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return filepath.EvalSymlinks(p)
	}
	defer xio.CloseIgnoringErrors(f)
	buf := make([]byte, darwinPathMax)
	if _, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), syscall.F_GETPATH,
		uintptr(unsafe.Pointer(&buf[0]))); errno != 0 {
		return filepath.EvalSymlinks(p)
	}
	resolved, _, _ := bytes.Cut(buf, []byte{0})
	return string(resolved), nil
}
