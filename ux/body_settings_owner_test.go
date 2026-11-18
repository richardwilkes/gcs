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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
)

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

// TestGlobalBodySettingsOwnerEditsTheDefaults verifies that the owner of the default body type edits the one the global
// settings hold, and resets it to the factory body type.
func TestGlobalBodySettingsOwnerEditsTheDefaults(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	t.Cleanup(gurps.SyncGlobalSheetSettings) // Runs after the swap below is undone.
	uxtest.SwapForTest(t, &global.Sheet, global.Sheet.Clone(nil))
	custom := gurps.FactoryBody()
	custom.Name = "Custom"
	globalBodySettings.SetBodySettings(custom)
	c.Equal("Custom", globalBodySettings.BodySettings(false).Name, "the defaults hold the body type set")
	c.Equal(gurps.FactoryBody().Name, globalBodySettings.BodySettings(true).Name,
		"a reset goes back to the factory body type")
}
