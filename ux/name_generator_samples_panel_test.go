// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestNameGeneratorSamplesPanel verifies that the samples panel shows ten names from a working definition, the reason
// when the definition cannot generate names, and that the samples follow edits and can be regenerated on demand.
func TestNameGeneratorSamplesPanel(t *testing.T) {
	c := check.New(t)
	d := newTestNameGeneratorEditorDockable(simpleNameGenerator("Alice"))
	c.NotNil(d.samples)
	c.Equal(strings.Repeat("Alice, ", 9)+"Alice", d.samples.label.text, "ten names, joined by commas")
	c.Equal(unison.DefaultLabelTheme.OnBackgroundInk, d.samples.label.ink, "names are drawn normally")

	trainingNamesPanel(d, d.model).removeOption(d.model.Entries[0])
	c.Contains(d.samples.label.text, "No training data has been provided", "the rebuilt samples say what is missing")
	c.Equal(unison.ThemeWarning, d.samples.label.ink, "the reason is drawn as a warning")

	d.undoMgr.Undo()
	c.Equal(strings.Repeat("Alice, ", 9)+"Alice", d.samples.label.text, "undo brings the names back")

	// Typing into a field is not a structural edit, yet the samples must follow it, which they do through Sync.
	stringFieldFor(t, d, d.model.Entries[0].KeyPrefix+"value").SetText("Bob")
	c.Equal(strings.Repeat("Bob, ", 9)+"Bob", d.samples.label.text, "the samples follow a typed change")

	d.model.Entries[0].Value = "Carol"
	button := buttonWithSVG(d.samples.AsPanel(), svg.Randomize)
	c.NotNil(button, "the header has a regenerate button")
	button.ClickCallback()
	c.Equal(strings.Repeat("Carol, ", 9)+"Carol", d.samples.label.text, "the button regenerates the samples")

	compound := gurps.NewNameGenerator()
	compound.Type = namegen.Compound
	d = newTestNameGeneratorEditorDockable(compound)
	c.Contains(d.samples.label.text, "at least one name generator", "an empty compound generator explains itself")
	c.NotContains(d.samples.label.text, "\n", "the message carries no stack trace")
}

// TestNameGeneratorSamplesPanelSyncKeepsSamplesOfAnUnchangedDefinition verifies that a sync that finds the definition
// as it was, such as one a field committing the value it already held brings about, leaves the samples alone rather
// than generating new ones. With 26 training names, new samples matching the old by chance is vanishingly unlikely.
func TestNameGeneratorSamplesPanelSyncKeepsSamplesOfAnUnchangedDefinition(t *testing.T) {
	c := check.New(t)
	names := make([]string, 0, 26)
	for letter := 'A'; letter <= 'Z'; letter++ {
		names = append(names, string(letter)+"ndor")
	}
	d := newTestNameGeneratorEditorDockable(simpleNameGenerator(names...))
	shown := d.samples.label.text
	d.samples.Sync()
	c.Equal(shown, d.samples.label.text)
}
