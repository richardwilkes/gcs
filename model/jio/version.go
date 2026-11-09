// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package jio

import (
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// These version numbers are used both as the version of the data files written to disk as well as the major version
// number for GCS library compatibility checking. The Go release of GCS is the first release to synchronize the data
// version and the library major version.
const (
	// CurrentDataVersion holds the current version for data files written with the current release. Note that this is
	// intentionally the same for all data files that GCS processes.
	CurrentDataVersion = 5
	// MinimumDataVersion holds the oldest version for data files that can be loaded, unless a file type passes
	// CheckHeader a minimum of its own.
	MinimumDataVersion    = 2
	MinimumLibraryVersion = 3
)

// Header holds the members every data and settings file starts with: the kind of data the file holds and the data
// version it was written with.
type Header struct {
	FileType string `json:"file_type"`
	Version  int    `json:"version"`
}

// NewHeader returns the header to write for a file holding the given kind of data.
func NewHeader(fileType string) Header {
	return Header{FileType: fileType, Version: CurrentDataVersion}
}

// CheckHeader returns an error if the header describes data this release cannot load: data written by a newer release,
// data of another kind than expectedFileType, or data older than minimum. minimum is MinimumDataVersion unless the
// kind's format changed after that, which raises it, or the kind's older files carry no version, which makes it 0.
// Data written by a newer release is reported as such whatever kind it claims, since that release may have renamed the
// kind. A header with no file type, as every file written before headers carried one has, is checked for its version
// alone.
func CheckHeader(h Header, expectedFileType string, minimum int) error {
	if h.Version > CurrentDataVersion {
		return errs.New(xstrings.Wrap("", i18n.Text("The data was written with a newer version of %[1]s and cannot be loaded. Please update %[1]s and try again.", xos.AppName), 76))
	}
	if h.FileType != "" && h.FileType != expectedFileType {
		return errs.New(xstrings.Wrap("", i18n.Text("The file holds %q data, but %q data was expected.", h.FileType, expectedFileType), 76))
	}
	if h.Version < minimum {
		return errs.New(xstrings.Wrap("", i18n.Text("The data was written with an older version of %s and cannot be loaded. You will need to load it with an earlier version that can read this version of the data and write the current format.", xos.AppName), 76))
	}
	return nil
}
