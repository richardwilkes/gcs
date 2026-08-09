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
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

var _ BodySettingsOwner = &bodySettingsOwnerForTest{}

// bodySettingsOwnerForTest is a minimal BodySettingsOwner standing in for a sheet.
type bodySettingsOwnerForTest struct {
	unison.Panel
	entity *gurps.Entity
	body   *gurps.Body
}

func (p *bodySettingsOwnerForTest) Entity() *gurps.Entity         { return p.entity }
func (p *bodySettingsOwnerForTest) BodySettingsTitle() string     { return "test" }
func (p *bodySettingsOwnerForTest) SetBodySettings(b *gurps.Body) { p.body = b }

func (p *bodySettingsOwnerForTest) BodySettings(_ bool) *gurps.Body { return p.body }

// TestBodySettingsDockableIdentifiedByOwner verifies that an open body settings dockable is matched to the settings it
// edits by the identity of its owner, so that a second request for the same settings reactivates the existing dockable
// rather than opening a duplicate.
func TestBodySettingsDockableIdentifiedByOwner(t *testing.T) {
	c := check.New(t)
	sheetOwner := &bodySettingsOwnerForTest{entity: gurps.NewEntity()}
	sheetOwner.Self = sheetOwner

	defaults := &bodySettingsDockable{owner: globalBodySettings}
	defaults.Self = defaults
	sheet := &bodySettingsDockable{owner: sheetOwner}
	sheet.Self = sheet

	c.True(isBodySettingsFor(defaults, globalBodySettings), "the defaults dockable must match the shared global owner")
	c.True(isBodySettingsFor(sheet, sheetOwner), "a sheet's dockable must match that sheet")
	c.False(isBodySettingsFor(defaults, sheetOwner), "the defaults dockable must not match a sheet")
	c.False(isBodySettingsFor(sheet, globalBodySettings), "a sheet's dockable must not match the defaults")

	// A dockable that isn't a body settings dockable never matches.
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	c.False(isBodySettingsFor(recorder, globalBodySettings), "an unrelated dockable must never match")
}

// TestGlobalBodySettingsOwnerIsShared verifies that only one globalBodySettingsOwner is ever created. The dockable for
// the default body settings is located by comparing owners, and globalBodySettingsOwner is a zero-size struct: the Go
// spec permits distinct zero-size variables to share an address, so a call site that allocates its own owner would be
// relying on unspecified behavior for "Default Body Type…" to reactivate the existing dockable instead of opening a
// duplicate, and would break outright if the struct ever gained a field.
func TestGlobalBodySettingsOwnerIsShared(t *testing.T) {
	c := check.New(t)
	c.NotNil(globalBodySettings, "the shared global body settings owner must exist")
	entries, err := os.ReadDir(".")
	c.NoError(err, "the package directory must be readable")
	fset := token.NewFileSet()
	var found []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		var file *ast.File
		if file, err = parser.ParseFile(fset, name, nil, 0); err != nil {
			c.NoError(err, "%s must be parsable", name)
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				if id, ok2 := lit.Type.(*ast.Ident); ok2 && id.Name == "globalBodySettingsOwner" {
					found = append(found, filepath.ToSlash(name))
				}
			}
			return true
		})
	}
	c.Equal([]string{"body_settings_owner.go"}, found,
		"globalBodySettingsOwner must be instantiated exactly once, by the shared globalBodySettings variable")
}
