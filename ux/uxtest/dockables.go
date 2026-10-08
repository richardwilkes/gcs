// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package uxtest

import (
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

// Closable is a dockable that can be closed as a tab, which is what the actions that open one hand back.
type Closable interface {
	unison.Dockable
	unison.TabCloser
}

// ActionForKey returns the action registered under the key binding ID, failing the test if there is none. The actions
// are registered when the menu bar is set up, so this is for use once StartHeadlessWorkspace has returned.
func ActionForKey(t *testing.T, key string) *unison.Action {
	t.Helper()
	for _, binding := range gurps.CurrentBindings() {
		if binding.ID == key {
			return binding.Action
		}
	}
	t.Fatalf("no action is registered under the key binding %q", key)
	return nil
}

// OpenedByAction executes action on the UI thread and returns the one dockable it opened, failing the test if it
// opened any other number of them.
func OpenedByAction(t *testing.T, screen *unison.HeadlessScreen, action *unison.Action) Closable {
	t.Helper()
	var opened []unison.Dockable
	screen.Do(func() {
		before := make(map[unison.Dockable]bool)
		for _, d := range workspace.AllDockables() {
			before[d] = true
		}
		action.Execute(nil)
		for _, d := range workspace.AllDockables() {
			if !before[d] {
				opened = append(opened, d)
			}
		}
	})
	if len(opened) != 1 {
		t.Fatalf("%s opened %d dockables; expected exactly one", action.Title, len(opened))
	}
	d, ok := opened[0].AsPanel().Self.(Closable)
	if !ok {
		t.Fatalf("%s opened a %T, which cannot be closed as a tab", action.Title, opened[0].AsPanel().Self)
	}
	return d
}

// SoleEditor returns the one open dockable that match accepts, as a T, failing the test if there is not exactly one or
// it is not a T. Only the lookup runs on the UI thread, so a caller reading anything from the editor does so in a
// screen.Do of its own afterwards.
func SoleEditor[T unison.Dockable](t *testing.T, screen *unison.HeadlessScreen, match func(unison.Dockable) bool) T {
	t.Helper()
	return onlyEditor[T](t, screen, match, nil)
}

// OtherEditor returns the one open dockable that match accepts besides d, as a T, failing the test if there is not
// exactly one or it is not a T; see SoleEditor.
func OtherEditor[T unison.Dockable](t *testing.T, screen *unison.HeadlessScreen, d T, match func(unison.Dockable) bool) T {
	t.Helper()
	return onlyEditor[T](t, screen, match, d)
}

// onlyEditor is what SoleEditor and OtherEditor share: it returns the one open dockable that match accepts, other than
// except when that is not nil, failing the test if there is not exactly one or it is not a T.
func onlyEditor[T unison.Dockable](t *testing.T, screen *unison.HeadlessScreen, match func(unison.Dockable) bool, except unison.Dockable) T {
	t.Helper()
	var editors []T
	screen.Do(func() {
		for _, one := range workspace.AllDockables() {
			if !match(one) {
				continue
			}
			if editor, ok := one.AsPanel().Self.(T); ok && (except == nil || editor.AsPanel() != except.AsPanel()) {
				editors = append(editors, editor)
			}
		}
	})
	if len(editors) != 1 {
		var zero T
		if except == nil {
			t.Fatalf("expected exactly one %T, found %d", zero, len(editors))
		}
		t.Fatalf("expected exactly one other %T, found %d", zero, len(editors))
	}
	return editors[0]
}

// CloseEditorWithoutPrompt closes a dockable that is expected to close without a prompt, failing the test if it is
// still open or a dialog came up. The close is posted rather than run through Do, since a prompt would be modal and Do
// would wait for it.
func CloseEditorWithoutPrompt(t *testing.T, screen *unison.HeadlessScreen, d Closable) {
	t.Helper()
	screen.Post(func() { d.AttemptClose() })
	screen.Sync()
	var stillOpen bool
	var windows int
	screen.Do(func() {
		stillOpen = slices.ContainsFunc(workspace.AllDockables(), func(open unison.Dockable) bool { return open.AsPanel().Self == d })
		windows = len(unison.Windows())
	})
	if stillOpen {
		t.Fatalf("%s is still open", d.Title())
	}
	if windows != 1 {
		t.Fatalf("closing %s left %d windows open; expected the workspace alone", d.Title(), windows)
	}
}
