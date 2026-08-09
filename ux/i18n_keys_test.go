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

// TestI18nTextArgsAreLiterals verifies that every i18n.Text call in the source tree is passed a single string literal.
// i18n.Text looks the whole string up in the translation catalog, so a call such as
// i18n.Text("Attributes: " + profile.Name) builds a key that varies per character and can never match an entry; worse,
// there is nothing constant for the extraction tooling to put in the catalog in the first place. The runtime portion
// belongs outside the lookup, as fmt.Sprintf(i18n.Text("Attributes: %s"), profile.Name).
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
			if len(call.Args) != 1 {
				c.Equal(1, len(call.Args), "%s:%d: i18n.Text must take exactly one argument", rel, pos.Line)
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			c.True(ok && lit.Kind == token.STRING,
				"%s:%d: i18n.Text must be given a string literal, so the lookup key is constant; move the runtime "+
					"portion into a fmt.Sprintf around it", rel, pos.Line)
			return true
		})
		return nil
	}), "the source tree must be walkable")
	c.True(checked > 100, "the walk should have found the i18n.Text calls, but only saw %d", checked)
}
