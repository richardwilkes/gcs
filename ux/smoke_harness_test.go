// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//go:build smoke

package ux

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
)

// The smoke tests start the whole application headless, exactly as main does apart from the handoff service, against a
// known set of fixtures, then drive it the way a user would and compare what they find against golden files. They are
// compiled only with the smoke build tag:
//
//	go test -tags smoke ./ux -run Smoke            # compare against the goldens
//	go test -tags smoke ./ux -run Smoke -update    # rewrite the goldens from what the tests see
//
// Everything a run could pick up from the machine it runs on is pinned, so that the same run produces the same pixels:
// the clock, the time zone, the application's name, version and copyright banner, the home directory, and the
// settings, which start from the factory defaults with the few changes smokeSettings makes.
//
// The fixtures live in testdata/smoke: master_library and user_library are the two libraries, and files holds
// individual documents, such as character sheets, that a test opens or saves. Each test works on its own copy of all
// three, made in a temporary directory, so no test can change the fixtures themselves or see what another test did.

// smokeNow is the time the application sees throughout a smoke test.
var smokeNow = time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC)

// smokeFixtureDir is where the smoke tests' fixtures are kept, relative to this package.
const smokeFixtureDir = "testdata/smoke"

// smokeSession is one running smoke test: the headless screen, the workspace window, and the directory holding the
// test's own copy of the fixtures.
type smokeSession struct {
	t      *testing.T
	c      check.Checker
	screen *unison.HeadlessScreen
	wnd    *unison.Window
	dir    string
	// beeps is how many beeps had been heard when the invariants were last checked.
	beeps int
	// logged holds the error-level log records written since the invariants were last checked. Anything may log, from
	// any goroutine, so it is guarded by logLock.
	logged  []string
	logLock sync.Mutex
	// reported holds the unnamed controls already reported, so that each is reported once however often it is seen.
	reported map[string]bool
}

// startSmoke pins the environment, copies the fixtures and starts the application with the named fixture files open,
// as though they had been given on the command line. The session is stopped, and everything it changed is put back,
// when the test ends. Anything the session records through Errors(), and any error the workspace would have reported in
// a dialog, fails the test.
func startSmoke(t *testing.T, files ...string) *smokeSession {
	t.Helper()
	s := &smokeSession{t: t, c: check.New(t), dir: t.TempDir(), reported: make(map[string]bool)}
	s.pinEnvironment()
	s.captureErrorLogs()
	for _, one := range []string{"master_library", "user_library", "files"} {
		s.c.NoError(os.CopyFS(filepath.Join(s.dir, one), os.DirFS(filepath.Join(smokeFixtureDir, one))), one)
	}
	s.useSmokeSettings()
	swapForTest(t, &Workspace, Workspace)
	RegisterKnownFileTypes()

	paths := make([]string, len(files))
	for i, one := range files {
		paths[i] = s.file(one)
	}
	screen, err := unison.StartHeadless(unison.HeadlessConfig{Width: 1400, Height: 900}, StartOptions(paths, false)...)
	if err != nil {
		t.Fatalf("unable to start the headless session: %v", err)
	}
	s.screen = screen
	// Registered ahead of Stop so that it runs after the session has ended, by which time everything it will record
	// has been recorded.
	t.Cleanup(func() {
		for _, one := range screen.Errors() {
			t.Errorf("the headless session recorded an error: %v", one)
		}
		s.checkLogs()
	})
	t.Cleanup(screen.Stop)
	// The navigator watches the libraries for changes and reloads itself a moment after one, on a timer. Left running,
	// the watches would see this test's copies of the fixtures being removed, and a reload still pending when the
	// session stops would fire anyway; either way this test's navigator would reload in whatever session the next test
	// has started, repopulating itself from that test's libraries and changing which of its rows are open. Registered
	// after Stop so that it runs before it, while the session can still act.
	t.Cleanup(func() {
		screen.Do(func() {
			for _, lib := range gurps.GlobalSettings().Libraries.List() {
				lib.StopAllWatches()
			}
		})
		s.settle()
	})
	screen.Do(func() {
		s.wnd = Workspace.Window
		Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }
	})
	if s.wnd == nil {
		t.Fatal("the workspace window was not created")
	}
	s.settle()
	return s
}

// captureErrorLogs records every error-level log record written while the test runs, for checkInvariants to report,
// while still writing everything to stderr. The records are passed on to a text handler of their own rather than the
// one being replaced, since that is slog's default handler, which writes through the log package, and once a handler of
// another kind is the default the log package writes through slog in turn, so the two would call each other forever.
func (s *smokeSession) captureErrorLogs() {
	prev := slog.Default()
	s.t.Cleanup(func() { slog.SetDefault(prev) })
	slog.SetDefault(slog.New(&smokeLogHandler{next: slog.NewTextHandler(os.Stderr, nil), s: s}))
}

// smokeLogHandler is the slog.Handler captureErrorLogs installs.
type smokeLogHandler struct {
	next slog.Handler
	s    *smokeSession
}

func (h *smokeLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= slog.LevelError || h.next.Enabled(ctx, level)
}

func (h *smokeLogHandler) Handle(ctx context.Context, r slog.Record) error { //nolint:gocritic // slog.Handler requires a value
	if r.Level >= slog.LevelError {
		var buf strings.Builder
		buf.WriteString(r.Message)
		r.Attrs(func(a slog.Attr) bool {
			fmt.Fprintf(&buf, " %s=%v", a.Key, a.Value)
			return true
		})
		h.s.logLock.Lock()
		h.s.logged = append(h.s.logged, buf.String())
		h.s.logLock.Unlock()
	}
	if h.next.Enabled(ctx, r.Level) {
		return h.next.Handle(ctx, r)
	}
	return nil
}

func (h *smokeLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &smokeLogHandler{next: h.next.WithAttrs(attrs), s: h.s}
}

func (h *smokeLogHandler) WithGroup(name string) slog.Handler {
	return &smokeLogHandler{next: h.next.WithGroup(name), s: h.s}
}

// checkInvariants checks what must hold at every step of every test, whatever the test is about, failing the test for
// each thing that does not: nothing in w may be a control without an accessible name (by the same rule as
// TestEveryControlHasAnAccessibleName), nothing may have beeped, and nothing may have been logged as an error. It runs
// whenever a test takes a GUI snapshot, so every screen a test looks at is checked.
// The where says which step it was, for the failure messages.
func (s *smokeSession) checkInvariants(where string, w *unison.Window) {
	s.t.Helper()
	if beeps := s.screen.Beeps(); beeps != s.beeps {
		s.t.Errorf("%s: the application beeped %d time(s)", where, beeps-s.beeps)
		s.beeps = beeps
	}
	s.checkLogs()
	tree := s.screen.AccessibilityTree(w)
	if tree == nil {
		return
	}
	tree.Walk(func(n *accessibility.Node) bool {
		if n.Ignored || n.Name != "" || !axNodeNeedsAName(n) {
			return true
		}
		desc := describeUnnamed(tree, n)
		if !s.reported[desc] {
			s.reported[desc] = true
			s.t.Errorf("%s: a control has no accessible name: %s", where, desc)
		}
		return true
	})
}

// checkLogs fails the test for every error-level record logged since it was last called.
func (s *smokeSession) checkLogs() {
	s.t.Helper()
	s.logLock.Lock()
	logged := s.logged
	s.logged = nil
	s.logLock.Unlock()
	for _, one := range logged {
		s.t.Errorf("an error was logged: %s", one)
	}
}

// describeUnnamed says where an unnamed node is: its role, the role and name of each named ancestor, and the text of
// whatever follows it, which is usually what a person would read as its label.
func describeUnnamed(tree *accessibility.Tree, n *accessibility.Node) string {
	var path []string
	for p := tree.Node(n.Parent); p != nil; p = tree.Node(p.Parent) {
		if !p.Ignored && p.Name != "" {
			path = append(path, fmt.Sprintf("%s %q", p.Role.Key(), p.Name))
		}
	}
	slices.Reverse(path)
	desc := n.Role.Key()
	if len(path) != 0 {
		desc += " in " + strings.Join(path, " > ")
	}
	if parent := tree.Node(n.Parent); parent != nil {
		for i, id := range parent.Children {
			if id != n.ID || i+1 >= len(parent.Children) {
				continue
			}
			if next := tree.Node(parent.Children[i+1]); next != nil {
				text := next.Name
				if text == "" && next.Text != nil {
					text = next.Text.Text
				}
				if text == "" {
					for _, cid := range next.Children {
						if child := tree.Node(cid); child != nil && child.Text != nil && child.Text.Text != "" {
							text = child.Text.Text
							break
						}
					}
				}
				if text != "" {
					desc += fmt.Sprintf(", before %q", text)
				}
			}
		}
	}
	return desc
}

// settle waits for the work the navigator defers to a timer to be done: the reload it asks for as it is created and
// whenever a library changes, and the resizing of its table. Sync cannot wait for these, since nothing can tell a timer
// that is about to fire from one that never will, but the navigator records that it has them pending. Starting up is
// not finished until they are done, since the reload can change what the navigator shows.
func (s *smokeSession) settle() {
	s.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		pending := false
		if !s.screen.Do(func() {
			n := Workspace.Navigator
			pending = n != nil && (n.needReload || n.adjustTableSizePending)
		}) || !pending {
			return
		}
		if time.Now().After(deadline) {
			s.t.Fatal("the navigator still has work pending after 5 seconds")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// pinEnvironment fixes everything outside the settings that the application would otherwise take from the machine or
// the moment it runs at.
func (s *smokeSession) pinEnvironment() {
	t := s.t
	swapForTest(t, &time.Local, time.UTC)
	swapForTest(t, &jio.Clock, func() time.Time { return smokeNow })
	swapForTest(t, &xos.AppName, "GCS")
	// Anything other than the test binary's own name, so that the Linux desktop integration done at startup skips
	// itself rather than installing icons for the test binary.
	swapForTest(t, &xos.AppCmdName, "gcs")
	swapForTest(t, &xos.AppVersion, "5.0.0")
	swapForTest(t, &xos.BuildNumber, "20260102150400")
	swapForTest(t, &xos.CopyrightStartYear, "1998")
	swapForTest(t, &xos.CopyrightEndYear, "2026")
	swapForTest(t, &xos.CopyrightHolder, "Richard A. Wilkes")
	// The caret blinks and tooltips appear on timers, which a capture could otherwise land either side of. With these,
	// a focused field always shows its caret and no tooltip appears unless a test hovers for half a minute.
	swapForTest(t, &unison.DefaultFieldTheme.BlinkRate, time.Hour)
	swapForTest(t, &unison.DefaultTooltipTheme.Delay, unison.DefaultTooltipTheme.Delay)
	swapForTest(t, &unison.DefaultTooltipTheme.Dismissal, unison.DefaultTooltipTheme.Dismissal)
	home := filepath.Join(s.dir, "home")
	s.c.NoError(os.MkdirAll(home, 0o750))
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
}

// useSmokeSettings replaces the global settings, for the length of the test, with the factory defaults plus the changes
// a smoke test needs, and points the libraries at the test's copies of the fixture libraries.
func (s *smokeSession) useSmokeSettings() {
	global := gurps.GlobalSettings()
	saved := *global
	s.t.Cleanup(func() {
		*global = saved
		applyGlobalSettings(global)
	})
	swapForTest(s.t, &gurps.SettingsPath, filepath.Join(s.dir, "settings.json"))
	*global = gurps.Settings{
		LastSeenGCSVersion: xos.AppVersion,
		General:            gurps.NewGeneralSettings(),
		Libraries:          library.NewLibraries(),
		Sheet:              gurps.FactorySheetSettings(),
	}
	// Nothing may reach the network.
	global.General.AppUpdateCheck = updatecheck.Never
	global.General.LibraryUpdateCheck = updatecheck.Never
	// A new sheet's description is otherwise filled in at random.
	global.General.AutoFillProfile = false
	// Each test starts from an empty workspace.
	global.General.RestoreWorkspaceOnStart = false
	// See pinEnvironment.
	global.General.TooltipDelay = gurps.TooltipDelayMax
	s.c.NoError(global.Libraries.Master().SetPath(filepath.Join(s.dir, "master_library")))
	s.c.NoError(global.Libraries.User().SetPath(filepath.Join(s.dir, "user_library")))
	global.EnsureValidity()
	applyGlobalSettings(global)
}

// applyGlobalSettings makes the given global settings take effect, as loading them at startup does.
func applyGlobalSettings(global *gurps.Settings) {
	gurps.SyncScriptExecTimeLimit()
	gurps.SyncGlobalSheetSettings()
	unison.SetThemeMode(global.ThemeMode)
	global.Colors.MakeCurrent()
	global.Fonts.MakeCurrent()
}

// file returns the path of the test's own copy of the named fixture file.
func (s *smokeSession) file(name string) string {
	return filepath.Join(s.dir, "files", name)
}
