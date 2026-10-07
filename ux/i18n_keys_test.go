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
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestI18nTextArgsAreLiterals verifies that every i18n.Text call in the source tree is passed a string literal as its
// first argument. i18n.Text looks that whole string up in the translation catalog, so a call such as
// i18n.Text("Attributes: " + profile.Name) builds a key that can never match an entry, and leaves the extraction
// tooling nothing constant to put in the catalog. The runtime portion belongs in the format arguments, as
// i18n.Text("Attributes: %s", profile.Name).
func TestI18nTextArgsAreLiterals(t *testing.T) {
	c := check.New(t)
	root, err := filepath.Abs("..")
	c.NoError(err, "the module root must be locatable")

	fileSet := token.NewFileSet()
	checked := 0
	c.NoError(filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip anything that isn't our own source.
			if name := d.Name(); path != root && (strings.HasPrefix(name, ".") || name == "testdata") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, parseErr := parser.ParseFile(fileSet, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Text" {
				return true
			}
			if pkg, ok2 := sel.X.(*ast.Ident); !ok2 || pkg.Name != "i18n" {
				return true
			}
			checked++
			pos := fileSet.Position(call.Pos())
			var first ast.Expr
			if len(call.Args) != 0 {
				first = call.Args[0]
			}
			lit, ok := first.(*ast.BasicLit)
			c.True(ok && lit.Kind == token.STRING,
				"%s:%d: i18n.Text must be given a string literal first, so the lookup key is constant; move the "+
					"runtime portion into its format arguments", rel, pos.Line)
			return true
		})
		return nil
	}), "the source tree must be walkable")
	c.True(checked > 100, "the walk should have found the i18n.Text calls, but only saw %d", checked)
}
