// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "io"

// LibraryUpdatePhase identifies the step a library update has reached, so the caller can say what is happening.
type LibraryUpdatePhase byte

const (
	// LibraryUpdateDownloading means the library's content is being retrieved.
	LibraryUpdateDownloading LibraryUpdatePhase = iota
	// LibraryUpdateInstalling means the retrieved content is being written into place.
	LibraryUpdateInstalling
)

// LibraryUpdateProgress reports how far a library update has gotten. fraction runs from 0 to 1 within the phase it is
// given, so a caller showing a progress bar should start the bar over whenever the phase changes. It is called often,
// and from more than one goroutine, though never from two at once, so it must be cheap and must not block; throttling
// and marshaling to a UI thread are the caller's business.
type LibraryUpdateProgress func(phase LibraryUpdatePhase, fraction float64)

// maxEstimatedFraction is as far as the download portion of the progress bar may advance while it is being measured
// against an estimate. Holding it short of the end keeps an estimate that turns out to be too small from showing a
// finished bar while data is still arriving.
const maxEstimatedFraction = 0.98

// estimatedFraction maps a running byte count onto the 0 to 1 range using a size that is only a guess. GitHub builds
// its source archives on demand and sends them without a Content-Length, and the git protocol doesn't announce the size
// of a pack either, so this is the best that can be done for the part of an update that is still arriving.
func estimatedFraction(received, estimate int64) float64 {
	if received <= 0 || estimate <= 0 {
		return 0
	}
	if fraction := float64(received) / float64(estimate); fraction < maxEstimatedFraction {
		return fraction
	}
	return maxEstimatedFraction
}

// exactFraction maps a running count onto the 0 to 1 range using a total that is known. A total of zero means there was
// nothing to do, which is as finished as it can be.
func exactFraction(current, total int64) float64 {
	switch {
	case current >= total:
		return 1
	case current <= 0:
		return 0
	default:
		return float64(current) / float64(total)
	}
}

// countingReader reports the size of each chunk of data that passes through it, so that a transfer whose total size
// isn't known in advance can still be measured. received is handed the size of the chunk, not a running total, since
// the git transport reads from more than one response and may do so from more than one goroutine.
type countingReader struct {
	r        io.Reader
	received func(n int64)
}

func (r *countingReader) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	if n > 0 && r.received != nil {
		r.received(int64(n))
	}
	return n, err
}

// countingReadCloser is a countingReader that closes the reader it was built from, for the case where what is being
// measured is an HTTP response body.
type countingReadCloser struct {
	countingReader
	closer io.Closer
}

func newCountingReadCloser(body io.ReadCloser, received func(n int64)) *countingReadCloser {
	return &countingReadCloser{
		countingReader: countingReader{r: body, received: received},
		closer:         body,
	}
}

func (r *countingReadCloser) Close() error {
	return r.closer.Close()
}
