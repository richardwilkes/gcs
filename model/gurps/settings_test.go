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
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// TestLoadSettingsOrDefaultsWithMissingFile verifies that the normal first-run case -- no settings file at all --
// yields factory defaults without logging an error or leaving a stray backup behind.
func TestLoadSettingsOrDefaultsWithMissingFile(t *testing.T) {
	c := check.New(t)
	count := countLogs(t, slog.LevelError)
	p := filepath.Join(t.TempDir(), "settings.json")
	settings := loadSettingsOrDefaults(p)
	c.NotNil(settings.General)
	c.NotNil(settings.Sheet)
	c.NotEqual(0, settings.Libraries.Len())
	c.Equal(int32(0), count.Load())
	_, err := os.Stat(p + ".bad")
	c.HasError(err)
}

func TestLoadSettingsOrDefaultsWithValidFile(t *testing.T) {
	c := check.New(t)
	count := countLogs(t, slog.LevelError)
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
	count := countLogs(t, slog.LevelError)
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

// TestLoadSettingsOrDefaultsWithNullLibraryEntry verifies that a settings file holding a null in place of a library is
// survivable. Decoding one used to panic, which bypassed this function's recovery entirely and crashed GCS at startup
// rather than leaving it with usable settings.
func TestLoadSettingsOrDefaultsWithNullLibraryEntry(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelError)
	p := filepath.Join(t.TempDir(), "settings.json")
	c.NoError(os.WriteFile(p, []byte(`{"last_seen_gcs_version":"1.2.3","libraries":{"a/b":null,`+
		`"someone/repo":{"title":"Good","path":"/libs/good"}}}`), 0o600))
	var settings Settings
	c.NotPanics(func() { settings = loadSettingsOrDefaults(p) })
	c.Equal("1.2.3", settings.LastSeenGCSVersion, "the rest of the settings file still loaded")
	c.NotNil(settings.Libraries.Lookup("someone/repo"), "the usable library entry survived")
	c.Nil(settings.Libraries.Lookup("a/b"), "the null entry was skipped")
}

// TestSettingsSaveWithNullMapEntries verifies that a settings file holding a JSON null in place of a navigator node, a
// column-sizing entry, or a PDF entry is survivable. Such entries decode without error, so the file loaded fine and
// then Save() -- which runs on quit and after any settings change -- panicked while pruning stale entries, losing the
// session's settings changes.
func TestSettingsSaveWithNullMapEntries(t *testing.T) {
	c := check.New(t)
	countLogs(t, slog.LevelError)
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	c.NoError(os.WriteFile(p, []byte(`{
	"library_explorer": {"nodes": {"F@bad": null, "F@good": {"id": "3abcdefghij", "last": 9999999999}}},
	"column_sizing": {"bad": null, "good": {"-1": 9999999, "0": 120}},
	"pdfs": {"bad": null, "good": {"last": 9999999999, "toc": {"Chapter 1": null}}}
}`), 0o600))
	settings := loadSettingsOrDefaults(p)
	c.Equal(2, len(settings.LibraryExplorer.Nodes), "both node entries decoded")
	c.Equal(2, len(settings.ColumnSizing), "both column sizing entries decoded")
	c.Equal(2, len(settings.PDFs), "both PDF entries decoded")

	prevPath := SettingsPath
	SettingsPath = filepath.Join(dir, "saved.json")
	t.Cleanup(func() { SettingsPath = prevPath })
	c.NotPanics(func() { c.NoError(settings.Save()) }, "saving must not panic on null entries")

	// The unusable entries are gone and the usable ones survived.
	c.Equal(1, len(settings.LibraryExplorer.Nodes), "the null node entry was dropped")
	c.NotNil(settings.LibraryExplorer.Nodes["F@good"], "the usable node entry survived")
	c.Equal(1, len(settings.ColumnSizing), "the null column sizing entry was dropped")
	c.Equal(float32(120), settings.ColumnSizing["good"][0], "the usable column sizing entry survived")
	c.Equal(1, len(settings.PDFs), "the null PDF entry was dropped")
	good := settings.PDFs["good"]
	c.NotNil(good, "the usable PDF entry survived")
	c.Equal(0, len(good.TOC), "the null table of contents entry was dropped")

	// What was written back is loadable and holds the same surviving entries.
	reloaded := loadSettingsOrDefaults(SettingsPath)
	c.Equal(1, len(reloaded.LibraryExplorer.Nodes), "the saved file has just the usable node entry")
	c.Equal(1, len(reloaded.ColumnSizing), "the saved file has just the usable column sizing entry")
	c.Equal(1, len(reloaded.PDFs), "the saved file has just the usable PDF entry")
}

// TestIDLookupsWithNullMapEntries verifies that the ID lookups treat a null settings entry as if it were absent rather
// than dereferencing it.
func TestIDLookupsWithNullMapEntries(t *testing.T) {
	c := check.New(t)
	settings := GlobalSettings()

	const nullPDF = "/does/not/exist/null.pdf"
	settings.PDFs[nullPDF] = nil
	t.Cleanup(func() { delete(settings.PDFs, nullPDF) })
	var id tid.TID
	c.NotPanics(func() { id = IDForPDFTOC(nullPDF, "Chapter 1", 3) })
	c.True(tid.IsKind(id, kinds.TableOfContents), "a usable TOC ID was returned")

	// A null in place of a title's page map is the same hazard one level down.
	const nullTOCPDF = "/does/not/exist/null_toc.pdf"
	settings.PDFs[nullTOCPDF] = &PDFInfo{TOC: map[string]map[int]tid.TID{"Chapter 1": nil}}
	t.Cleanup(func() { delete(settings.PDFs, nullTOCPDF) })
	c.NotPanics(func() { id = IDForPDFTOC(nullTOCPDF, "Chapter 1", 3) })
	c.True(tid.IsKind(id, kinds.TableOfContents), "a usable TOC ID was returned")

	const nullNode = "/does/not/exist/null_node"
	if settings.LibraryExplorer.Nodes == nil {
		settings.LibraryExplorer.Nodes = make(map[string]*NavNodeInfo)
	}
	settings.LibraryExplorer.Nodes[nullNode] = nil
	t.Cleanup(func() { delete(settings.LibraryExplorer.Nodes, nullNode) })
	c.NotPanics(func() { id = IDForNavNode(nullNode, kinds.NavigatorFile) })
	c.True(tid.IsKind(id, kinds.NavigatorFile), "a usable navigator node ID was returned")
}

// TestSetAsideDamagedSettingsWithMissingFile verifies that setting aside a file that isn't there is a no-op rather than
// an error log.
func TestSetAsideDamagedSettingsWithMissingFile(t *testing.T) {
	c := check.New(t)
	count := countLogs(t, slog.LevelError)
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
	count := countLogs(t, slog.LevelError)
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

// legacyOpenSellMoneyEnv names the settings file the child process of TestGlobalSettingsLoadWithLegacyOpenSellMoney
// loads.
const legacyOpenSellMoneyEnv = "GCS_TEST_LEGACY_OPEN_SELL_MONEY"

// TestGlobalSettingsLoadWithLegacyOpenSellMoney checks that loading the global settings doesn't hang on sell money
// holding legacy-format data with an open flag, which the unmarshal records as a closed state from inside
// GlobalSettings' initializer, where calling GlobalSettings() again would deadlock (see closedStateMap). The load runs
// in a child process, since this process's global settings are already loaded.
func TestGlobalSettingsLoadWithLegacyOpenSellMoney(t *testing.T) {
	c := check.New(t)
	if path := os.Getenv(legacyOpenSellMoneyEnv); path != "" {
		SettingsPath = path
		settings := GlobalSettings()
		c.Equal(1, len(settings.Sheet.SellMoney), "the coin is kept and the legacy container dropped")
		if len(settings.Sheet.SellMoney) == 1 {
			c.Equal("Coin", settings.Sheet.SellMoney[0].Name)
		}
		return
	}
	path := filepath.Join(t.TempDir(), "settings.json")
	data := fmt.Sprintf(`{"sheet_settings":{"sell_money":[
	{"id":%q,"description":"Coin","tags":["Money"],
	 "modifiers":[{"type":"eqp_modifier_container","name":"Group","open":true}]},
	{"type":"equipment_container","description":"Purse","tags":["Money"],"open":true}
]}}`, tid.MustNewTID(kinds.Equipment))
	c.NoError(os.WriteFile(path, []byte(data), 0o600))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	//nolint:gosec // Re-executing this test binary, so that the global settings are loaded afresh
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run", "^TestGlobalSettingsLoadWithLegacyOpenSellMoney$", "-test.v")
	cmd.Env = append(os.Environ(), legacyOpenSellMoneyEnv+"="+path)
	out, err := cmd.CombinedOutput()
	c.NoError(ctx.Err(), "loading the settings must not hang")
	c.NoError(err, string(out))
	c.True(strings.Contains(string(out), "--- PASS"), string(out))
}

// TestClosedStateDuringSettingsInitialization checks that closed states set on the goroutine loading the global
// settings are kept aside and folded into them once loaded (see closedStateMap), while other goroutines keep using the
// loaded states.
func TestClosedStateDuringSettingsInitialization(t *testing.T) {
	c := check.New(t)
	settings := GlobalSettings()
	const key = "n:closed-during-init-test"
	t.Cleanup(func() {
		initializingGoroutine.Store(0)
		SetClosedState(key, false)
	})
	initializingGoroutine.Store(currentGoroutineID())
	c.True(SetClosedState(key, true))
	c.True(IsClosed(key), "the state is seen on the initializing goroutine")
	closedLock.Lock()
	_, loaded := settings.Closed[key]
	closedLock.Unlock()
	c.False(loaded, "the loaded settings are left alone while loading")
	var seenElsewhere bool
	done := make(chan struct{})
	go func() {
		seenElsewhere = IsClosed(key)
		close(done)
	}()
	<-done
	c.False(seenElsewhere, "another goroutine reads the loaded settings")
	initializingGoroutine.Store(0)
	settings.foldClosedDuringInit()
	c.True(IsClosed(key), "the state is folded into the settings once loaded")
	c.Equal(0, len(closedDuringInit))
}

// TestLoadSettingsKeepsTheFileWhenASellMoneyEntryIsUnreadable checks that an entry of the sell money that can't be
// decoded is dropped, with the rest loaded, rather than failing the whole load, which would set the user's settings
// file aside and start from factory defaults over one coin.
func TestLoadSettingsKeepsTheFileWhenASellMoneyEntryIsUnreadable(t *testing.T) {
	c := check.New(t)
	path := filepath.Join(t.TempDir(), "settings.json")
	data := fmt.Sprintf(`{"general":{"default_player_name":"Keep Me"},"sheet_settings":{"sell_money":[
	{"id":%q,"description":"Bad Coin","tags":["Money"],"quantity":"abc"},
	{"id":%q,"description":"Copper Penny","tags":["Money"],"base_value":"1"}
]}}`, tid.MustNewTID(kinds.Equipment), tid.MustNewTID(kinds.Equipment))
	c.NoError(os.WriteFile(path, []byte(data), 0o600))
	settings := loadSettingsOrDefaults(path)
	c.Equal("Keep Me", settings.General.DefaultPlayerName, "the rest of the settings are loaded")
	c.Equal(1, len(settings.Sheet.SellMoney), "the unreadable entry is dropped")
	if len(settings.Sheet.SellMoney) == 1 {
		c.Equal("Copper Penny", settings.Sheet.SellMoney[0].Name)
	}
	c.True(xos.FileExists(path), "the settings file is left where it was")
	c.False(xos.FileExists(path+".bad"), "and not set aside")
}
