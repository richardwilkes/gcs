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
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// progressStep is one call made to a LibraryUpdateProgress.
type progressStep struct {
	phase    LibraryUpdatePhase
	fraction float64
}

// recordProgress returns a reporter that collects everything it is told, along with the slice it collects into. The
// download runs on the calling goroutine, so no synchronization is needed here.
func recordProgress() (LibraryUpdateProgress, *[]progressStep) {
	var steps []progressStep
	return func(phase LibraryUpdatePhase, fraction float64) {
		steps = append(steps, progressStep{phase: phase, fraction: fraction})
	}, &steps
}

// lastFraction returns the final fraction reported for the given phase, or -1 if the phase was never reported.
func lastFraction(steps []progressStep, phase LibraryUpdatePhase) float64 {
	fraction := -1.0
	for _, step := range steps {
		if step.phase == phase {
			fraction = step.fraction
		}
	}
	return fraction
}

// maxFraction returns the largest fraction reported for the given phase, or -1 if the phase was never reported.
func maxFraction(steps []progressStep, phase LibraryUpdatePhase) float64 {
	fraction := -1.0
	for _, step := range steps {
		if step.phase == phase && step.fraction > fraction {
			fraction = step.fraction
		}
	}
	return fraction
}

// buildLibraryArchive produces a GitHub-style source archive: everything below a single top-level directory named for
// the repository and the commit, with the library's content under that directory's "Library" folder.
func buildLibraryArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)
	for name, content := range files {
		f, err := w.Create("someone-repo-abc1234/" + name)
		if err != nil {
			t.Fatalf("unable to add %s to the archive: %v", name, err)
		}
		if _, err = f.Write([]byte(content)); err != nil {
			t.Fatalf("unable to write %s into the archive: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("unable to close the archive: %v", err)
	}
	return buffer.Bytes()
}

// TestLibraryDownloadReportsProgress verifies that a download reports progress that a bar can actually be driven from:
// the two phases in order, fractions that never go backwards within a phase, and an install phase that ends at its end.
// Before this, the library update had nothing to report and the window showed an indeterminate bar throughout.
func TestLibraryDownloadReportsProgress(t *testing.T) {
	c := check.New(t)
	archive := buildLibraryArchive(t, map[string]string{
		"README.md":             "ignored, since it isn't under Library",
		"Library/one.gct":       strings.Repeat("a", 4096),
		"Library/sub/two.gct":   strings.Repeat("b", 8192),
		"Elsewhere/three.gct":   "ignored, since it isn't under Library",
		"Library/sub/three.gct": strings.Repeat("c", 2048),
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(archive) //nolint:errcheck // Nothing useful can be done with a failure to write to the test client
	}))
	defer srv.Close()
	dir := filepath.Join(t.TempDir(), "lib")
	lib := NewLibrary("Test", "someone", "", "repo", dir)
	release := Release{Version: "1.0.0", ZipFileURL: srv.URL}

	progress, steps := recordProgress()
	c.NoError(lib.Download(t.Context(), srv.Client(), &release, progress))

	// The content must have made it to disk, since progress that doesn't describe real work is worse than none.
	data, err := os.ReadFile(filepath.Join(dir, "sub", "two.gct"))
	c.NoError(err)
	c.Equal(8192, len(data))
	_, err = os.Stat(filepath.Join(dir, "README.md"))
	c.True(os.IsNotExist(err), "only the content under Library belongs in the library")

	c.True(len(*steps) > 1, "the update must report more than its completion")
	seenInstalling := false
	last := map[LibraryUpdatePhase]float64{}
	for _, step := range *steps {
		c.True(step.fraction >= 0 && step.fraction <= 1, "fraction %v must be within 0 to 1", step.fraction)
		if step.phase == LibraryUpdateInstalling {
			seenInstalling = true
		} else {
			c.False(seenInstalling, "the phases must be reported in order")
			c.True(step.fraction <= maxEstimatedFraction,
				"an estimated fraction must stay short of the end, was %v", step.fraction)
		}
		c.True(step.fraction >= last[step.phase], "%v went backwards: %v after %v", step.phase, step.fraction,
			last[step.phase])
		last[step.phase] = step.fraction
	}
	c.True(seenInstalling, "the install phase must be reported")
	c.Equal(1.0, lastFraction(*steps, LibraryUpdateInstalling), "the install phase must finish at its end")
}

// TestLibraryDownloadRecordsSize verifies that a download records how much it transferred and that the next one scales
// its progress against that rather than against the built-in guess. GitHub sends these archives without a
// Content-Length, so the recorded size is the only thing that makes the download portion of the bar meaningful.
func TestLibraryDownloadRecordsSize(t *testing.T) {
	c := check.New(t)
	archive := buildLibraryArchive(t, map[string]string{"Library/one.gct": strings.Repeat("a", 4096)})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(archive) //nolint:errcheck // Nothing useful can be done with a failure to write to the test client
	}))
	defer srv.Close()
	dir := filepath.Join(t.TempDir(), "lib")
	lib := NewLibrary("Test", "someone", "", "repo", dir)
	release := Release{Version: "1.0.0", ZipFileURL: srv.URL}

	// The first download has nothing to go on but the built-in guess, which is far larger than this archive, so the
	// download portion of the bar barely moves.
	progress, steps := recordProgress()
	c.NoError(lib.Download(t.Context(), srv.Client(), &release, progress))
	c.True(maxFraction(*steps, LibraryUpdateDownloading) < 0.5,
		"a guess this far off should not have produced a nearly complete bar")

	// The version must still be the first thing in the file, since that is all VersionOnDisk() reads, and the size must
	// follow it.
	data, err := os.ReadFile(filepath.Join(dir, releaseFile))
	c.NoError(err)
	lines := strings.Split(string(data), "\n")
	c.True(len(lines) >= 2, "the release file must hold both the version and the size")
	c.Equal("1.0.0", lines[0])
	c.Equal("1.0.0", lib.VersionOnDisk())
	c.Equal(int64(len(archive)), lib.recordedDownloadSize())

	// The second download knows what the first one cost, so its bar can run all the way to the clamp.
	release.Version = "1.1.0"
	progress, steps = recordProgress()
	c.NoError(lib.Download(t.Context(), srv.Client(), &release, progress))
	c.Equal(maxEstimatedFraction, maxFraction(*steps, LibraryUpdateDownloading),
		"a download measured against its own recorded size should reach the clamp")
	c.Equal(int64(len(archive)), lib.recordedDownloadSize())
}

// seedLibrary creates a library directory holding one file and a release file naming version 0.9.0, so that a failed or
// canceled update has something recognizable to have put back.
func seedLibrary(t *testing.T, dir string) *Library {
	t.Helper()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("unable to create the library directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "old.gct"), []byte("the previous content"), 0o640); err != nil {
		t.Fatalf("unable to write the previous content: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, releaseFile), []byte("0.9.0\n1024\n"), 0o640); err != nil {
		t.Fatalf("unable to write the previous release file: %v", err)
	}
	return NewLibrary("Test", "someone", "", "repo", dir)
}

// checkLibraryUntouched verifies that the library still holds what seedLibrary() put there.
func checkLibraryUntouched(c check.Checker, lib *Library, dir string) {
	data, err := os.ReadFile(filepath.Join(dir, "old.gct"))
	c.NoError(err, "the previous content must be back in place")
	c.Equal("the previous content", string(data))
	c.Equal("0.9.0", lib.VersionOnDisk(), "the library must still report the version it actually holds")
}

// TestLibraryDownloadCanceledWhileInstalling verifies that stopping an update part way through unpacking leaves the
// library holding what it held before. The Cancel button in the update window is what reaches this, and it would be
// worse than useless if using it could leave a half-written library behind.
func TestLibraryDownloadCanceledWhileInstalling(t *testing.T) {
	c := check.New(t)
	archive := buildLibraryArchive(t, map[string]string{
		"Library/one.gct":   strings.Repeat("a", 4096),
		"Library/two.gct":   strings.Repeat("b", 4096),
		"Library/three.gct": strings.Repeat("c", 4096),
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(archive) //nolint:errcheck // Nothing useful can be done with a failure to write to the test client
	}))
	defer srv.Close()
	dir := filepath.Join(t.TempDir(), "lib")
	lib := seedLibrary(t, dir)
	release := Release{Version: "1.0.0", ZipFileURL: srv.URL}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	err := lib.Download(ctx, srv.Client(), &release, func(phase LibraryUpdatePhase, _ float64) {
		if phase == LibraryUpdateInstalling {
			cancel()
		}
	})
	c.HasError(err, "a canceled update must be reported as a failure")
	c.True(errors.Is(err, context.Canceled), "the failure must be recognizable as the cancellation: %v", err)
	checkLibraryUntouched(c, lib, dir)
	_, err = os.Stat(filepath.Join(dir, "one.gct"))
	c.True(os.IsNotExist(err), "nothing from the canceled update may be left behind")
}

// TestLibraryDownloadCanceledWhileDownloading verifies the same for a cancellation during the transfer, which is where
// most of the waiting -- and therefore most of the canceling -- happens.
func TestLibraryDownloadCanceledWhileDownloading(t *testing.T) {
	c := check.New(t)
	archive := buildLibraryArchive(t, map[string]string{"Library/one.gct": strings.Repeat("a", 64*1024)})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(archive[:len(archive)/2]) //nolint:errcheck // Nothing useful can be done with a failure to write here
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush() // Without this, nothing reaches the client until the handler returns, which it never does
		}
		<-r.Context().Done() // The rest never arrives, so the client is left waiting until it gives up
	}))
	defer srv.Close()
	dir := filepath.Join(t.TempDir(), "lib")
	lib := seedLibrary(t, dir)
	release := Release{Version: "1.0.0", ZipFileURL: srv.URL}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	err := lib.Download(ctx, srv.Client(), &release, func(phase LibraryUpdatePhase, _ float64) {
		if phase == LibraryUpdateDownloading {
			cancel()
		}
	})
	c.HasError(err, "a canceled download must be reported as a failure")
	checkLibraryUntouched(c, lib, dir)
}

// TestRecordedDownloadSizeToleratesJunk verifies that a release file without a usable size is reported as "not known"
// rather than as an error or a nonsense estimate. Libraries written by older versions of GCS hold only a version, and
// the file is plain text sitting in the user's library directory, so anything at all may be in it.
func TestRecordedDownloadSizeToleratesJunk(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	lib := NewLibrary("Test", "someone", "", "repo", dir)
	releasePath := filepath.Join(dir, releaseFile)
	for _, content := range []string{"1.0.0\n", "1.0.0", "1.0.0\nnot a number\n", "1.0.0\n-12\n", "1.0.0\n\n", ""} {
		c.NoError(os.WriteFile(releasePath, []byte(content), 0o600))
		c.Equal(int64(0), lib.recordedDownloadSize(), "content %q", content)
	}
	c.NoError(os.WriteFile(releasePath, []byte("1.0.0\n4096\ntrailing junk\n"), 0o600))
	c.Equal(int64(4096), lib.recordedDownloadSize())
	c.NoError(os.Remove(releasePath))
	c.Equal(int64(0), lib.recordedDownloadSize(), "a missing file")
}

// TestEstimatedFraction verifies the mapping used while the size of what is arriving is only a guess.
func TestEstimatedFraction(t *testing.T) {
	c := check.New(t)
	c.Equal(0.0, estimatedFraction(0, 100))
	c.Equal(0.0, estimatedFraction(-1, 100), "a negative count cannot move the bar")
	c.Equal(0.0, estimatedFraction(50, 0), "without an estimate there is nothing to measure against")
	c.Equal(0.25, estimatedFraction(25, 100))
	c.Equal(maxEstimatedFraction, estimatedFraction(100, 100), "a perfect guess still stops short of the end")
	c.Equal(maxEstimatedFraction, estimatedFraction(1000, 100), "an undersized guess must not overrun the end")
}

// TestExactFraction verifies the mapping used once the total is known.
func TestExactFraction(t *testing.T) {
	c := check.New(t)
	c.Equal(0.0, exactFraction(0, 100))
	c.Equal(0.0, exactFraction(-1, 100))
	c.Equal(0.5, exactFraction(50, 100))
	c.Equal(1.0, exactFraction(100, 100))
	c.Equal(1.0, exactFraction(200, 100), "an overrun is still just finished")
	c.Equal(1.0, exactFraction(0, 0), "nothing to do is as finished as it gets")
}

// TestCountingReaderReportsEveryChunk verifies that the reader used to measure a transfer accounts for all of it, since
// the count it produces both drives the bar and becomes the next update's estimate.
func TestCountingReaderReportsEveryChunk(t *testing.T) {
	c := check.New(t)
	var total int64
	var calls int
	r := &countingReader{
		r: bytes.NewReader(bytes.Repeat([]byte("x"), 1000)),
		received: func(n int64) {
			total += n
			calls++
		},
	}
	buffer := make([]byte, 300)
	for {
		if _, err := r.Read(buffer); err != nil {
			break
		}
	}
	c.Equal(int64(1000), total)
	c.Equal(4, calls, "an empty read must not be reported")
}
