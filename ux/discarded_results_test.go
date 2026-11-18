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
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestNoDiscardedConstructorResults verifies that no statement in the source tree is nothing but a call to another
// package's New* constructor. Such a call is a leftover that neither the compiler nor the linter flags, yet it still
// allocates every time it runs: one in the toggle cell builder once cost a CheckBox panel per toggle cell per rebuild.
// Only qualified calls are checked, since a New* function in this module may return nothing, as NewSheetFromTemplate
// does.
func TestNoDiscardedConstructorResults(t *testing.T) {
	c := check.New(t)
	files := 0
	var discarded []string
	walkModuleSources(c, func(fileSet *token.FileSet, rel string, file *ast.File) {
		files++
		pkgNames := importedPackageNames(file)
		ast.Inspect(file, func(node ast.Node) bool {
			stmt, ok := node.(*ast.ExprStmt)
			if !ok {
				return true
			}
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !strings.HasPrefix(sel.Sel.Name, "New") {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || !pkgNames[pkg.Name] {
				return true
			}
			discarded = append(discarded, fmt.Sprintf("%s:%d: %s.%s", rel, fileSet.Position(call.Pos()).Line,
				pkg.Name, sel.Sel.Name))
			return true
		})
	})
	c.Equal(0, len(discarded), "these constructors have their result discarded; remove the call or keep what it "+
		"builds:\n%s", strings.Join(discarded, "\n"))
	c.True(files > 100, "the walk should have found the source files, but only saw %d", files)
}

// importedPackageNames returns the names the file's imports are referred to by, so that a call such as unison.NewLabel
// can be told apart from a method call on a variable.
func importedPackageNames(file *ast.File) map[string]bool {
	names := make(map[string]bool, len(file.Imports))
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name != "_" && imp.Name.Name != "." {
				names[imp.Name.Name] = true
			}
			continue
		}
		if path, err := strconv.Unquote(imp.Path.Value); err == nil {
			names[path[strings.LastIndexByte(path, '/')+1:]] = true
		}
	}
	return names
}
