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
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// countErrorLogging installs a slog handler that counts error-level records for the duration of the test and returns
// the counter.
func countErrorLogging(t *testing.T) *atomic.Int32 {
	t.Helper()
	var count atomic.Int32
	prev := slog.Default()
	slog.SetDefault(slog.New(errorCountingHandler{count: &count}))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &count
}

// TestLoadSettingsOrDefaultsWithMissingFile verifies that the normal first-run case -- no settings file at all -- yields
// factory defaults without logging an error or leaving a stray backup behind.
func TestLoadSettingsOrDefaultsWithMissingFile(t *testing.T) {
	c := check.New(t)
	count := countErrorLogging(t)
	p := filepath.Join(t.TempDir(), "settings.json")
	settings := loadSettingsOrDefaults(p)
	c.NotNil(settings.General)
	c.NotNil(settings.Sheet)
	c.NotEqual(0, len(settings.LibrarySet))
	c.Equal(int32(0), count.Load())
	_, err := os.Stat(p + ".bad")
	c.HasError(err)
}

// TestLoadSettingsOrDefaultsWithValidFile verifies that a usable settings file is loaded rather than discarded.
func TestLoadSettingsOrDefaultsWithValidFile(t *testing.T) {
	c := check.New(t)
	count := countErrorLogging(t)
	p := filepath.Join(t.TempDir(), "settings.json")
	c.NoError(os.WriteFile(p, []byte(`{"last_seen_gcs_version":"1.2.3","deep_search":["alpha"]}`), 0o600))
	settings := loadSettingsOrDefaults(p)
	c.Equal("1.2.3", settings.LastSeenGCSVersion)
	c.Equal([]string{"alpha"}, settings.DeepSearch)
	c.Equal(int32(0), count.Load())
	_, err := os.Stat(p + ".bad")
	c.HasError(err)
}

// TestLoadSettingsOrDefaultsWithCorruptFile verifies that a settings file that exists but can't be loaded is reported
// and set aside rather than silently discarded. Prior to this, the load error was thrown away, so a corrupt file was
// indistinguishable from a missing one and the next Save quietly overwrote whatever the user had.
func TestLoadSettingsOrDefaultsWithCorruptFile(t *testing.T) {
	c := check.New(t)
	count := countErrorLogging(t)
	p := filepath.Join(t.TempDir(), "settings.json")
	const corrupt = `{"last_seen_gcs_version":"1.2.3",`
	c.NoError(os.WriteFile(p, []byte(corrupt), 0o600))
	settings := loadSettingsOrDefaults(p)

	// Factory defaults are used, and the failure was reported.
	c.Equal(xos.AppVersion, settings.LastSeenGCSVersion)
	c.NotNil(settings.General)
	c.NotNil(settings.Sheet)
	c.NotEqual(int32(0), count.Load())

	// The damaged file was moved aside intact, so a subsequent Save can't overwrite it.
	_, err := os.Stat(p)
	c.HasError(err)
	data, err := os.ReadFile(p + ".bad")
	c.NoError(err)
	c.Equal(corrupt, string(data))
}

// TestSetAsideDamagedSettingsWithMissingFile verifies that setting aside a file that isn't there is a no-op rather than
// an error log.
func TestSetAsideDamagedSettingsWithMissingFile(t *testing.T) {
	c := check.New(t)
	count := countErrorLogging(t)
	p := filepath.Join(t.TempDir(), "settings.json")
	setAsideDamagedSettings(p)
	c.Equal(int32(0), count.Load())
	_, err := os.Stat(p + ".bad")
	c.HasError(err)
}

// TestSetAsideDamagedSettingsWithExistingBackup verifies that a backup left over from an earlier failure isn't
// clobbered: the new backup gets a timestamp added to its name instead.
func TestSetAsideDamagedSettingsWithExistingBackup(t *testing.T) {
	c := check.New(t)
	count := countErrorLogging(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	const older = `older damaged settings`
	c.NoError(os.WriteFile(p+".bad", []byte(older), 0o600))
	const newer = `newer damaged settings`
	c.NoError(os.WriteFile(p, []byte(newer), 0o600))

	setAsideDamagedSettings(p)
	c.Equal(int32(0), count.Load())

	// The original backup is untouched and the settings file was moved out of the way.
	data, err := os.ReadFile(p + ".bad")
	c.NoError(err)
	c.Equal(older, string(data))
	_, err = os.Stat(p)
	c.HasError(err)

	// The newer copy landed in a timestamped sibling.
	c.Equal(newer, readOnlyTimestampedBackup(t, dir))
}

// readOnlyTimestampedBackup returns the content of the sole timestamped ".bad" file in the given directory, failing the
// test if there isn't exactly one.
func readOnlyTimestampedBackup(t *testing.T, dir string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "settings.json.*.bad"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one timestamped backup, got %v", matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
