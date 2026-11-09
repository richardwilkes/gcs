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

// These numbers govern which data files a release can load and which releases can load the files it writes.
// CurrentDataVersion is written into every data file and is also the newest major version of library releases that
// this release accepts. MinimumReaderVersion is written alongside it, so that an older release can tell whether a file
// from a newer one is still safe for it to load. MinimumDataVersion is the oldest data version that can be loaded.
const (
	// CurrentDataVersion holds the current version for data files written with the current release. Note that this is
	// intentionally the same for all data files that GCS processes.
	CurrentDataVersion = 6
	// MinimumReaderVersion holds the CurrentDataVersion of the oldest release that can load what this release writes
	// without losing meaning. Raise it only for a breaking change.
	MinimumReaderVersion = 6
	// MinimumDataVersion holds the oldest version for data files that can be loaded, unless a file type sets a higher
	// minimum through CheckHeaderWithMinimum.
	MinimumDataVersion = 2
	// MinimumLibraryVersion holds the oldest major version of library releases that this release accepts.
	MinimumLibraryVersion = 3
)

// Header holds the members every top-level data file starts with.
type Header struct {
	Type       string `json:"type,omitzero"`
	Version    int    `json:"version"`
	MinVersion int    `json:"min_version,omitzero"`
}

// NewHeader returns the header to write for a data file of the given kind.
func NewHeader(kind string) Header {
	return Header{Type: kind, Version: CurrentDataVersion, MinVersion: MinimumReaderVersion}
}

// CheckHeader returns an error if the header describes data this release cannot load. 'newer' is true when the data
// was written by a newer release, but one that declared it still readable by this release.
func CheckHeader(h Header, expectedType string) (newer bool, err error) {
	return CheckHeaderWithMinimum(h, expectedType, MinimumDataVersion)
}

// CheckHeaderWithMinimum is like CheckHeader, but for data whose format changed after MinimumDataVersion, so that files
// older than the given minimum are refused rather than loaded. A type mismatch is only refused for data versions that
// always write their type, and only when both the header and 'expectedType' name one.
func CheckHeaderWithMinimum(h Header, expectedType string, minimum int) (newer bool, err error) {
	if h.MinVersion > CurrentDataVersion || (h.Version > CurrentDataVersion && h.MinVersion == 0) {
		return false, errs.New(newerVersionMessage())
	}
	if h.Version < minimum {
		return false, errs.New(xstrings.Wrap("", i18n.Text("The data was written with an older version of %s and cannot be loaded. You will need to load it with an earlier version that can read this version of the data and write the current format.", xos.AppName), 76))
	}
	if h.Version >= typedDataVersion && h.Type != "" && expectedType != "" && h.Type != expectedType {
		return false, errs.New(xstrings.Wrap("", i18n.Text("The file holds %q data, but %q data was expected.", h.Type, expectedType), 76))
	}
	return h.Version > CurrentDataVersion, nil
}

// typedDataVersion is the first data version whose files always carry their type.
const typedDataVersion = 6

func newerVersionMessage() string {
	return xstrings.Wrap("", i18n.Text("The data was written with a newer version of %[1]s and cannot be loaded. Please update %[1]s and try again.", xos.AppName), 76)
}
