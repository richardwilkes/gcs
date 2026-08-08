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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestKeyBindingIDsAreUnique verifies that no two actions are registered with the same key binding ID.
// gurps.RegisterKeyBinding silently ignores a duplicate ID, so an action that reuses one is never added to the binding
// set: it can't be seen or assigned a binding in the Menu Keys settings, KeyBindings.MakeCurrent() never updates it,
// and any binding the user assigns to that ID applies only to the action that claimed it first.
func TestKeyBindingIDsAreUnique(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	ids := keyBindingIDsInSource(c)
	c.True(len(ids) > 1, "actions.go must contain key binding registrations")
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		c.False(seen[id], "key binding ID %q is registered more than once; only the first registration takes effect", id)
		seen[id] = true
	}
	registered := make(map[string]bool, len(ids))
	for _, one := range gurps.CurrentBindings() {
		registered[one.ID] = true
	}
	for id := range seen {
		c.True(registered[id], "key binding ID %q was never added to the binding set", id)
	}
}

// TestEquipmentLibraryActionsAreSeparatelyBindable verifies that the two similarly-named equipment library actions each
// have their own key binding, since they were once registered with the same ID.
func TestEquipmentLibraryActionsAreSeparatelyBindable(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	byAction := make(map[*unison.Action]string)
	for _, one := range gurps.CurrentBindings() {
		byAction[one.Action] = one.ID
	}
	equipmentID, ok := byAction[newEquipmentLibraryAction]
	c.True(ok, "New Equipment Library must have a key binding")
	modifiersID, ok := byAction[newEquipmentModifiersLibraryAction]
	c.True(ok, "New Equipment Modifiers Library must have a key binding")
	c.NotEqual(equipmentID, modifiersID, "the two equipment library actions must not share a key binding ID")
}

// keyBindingIDsInSource returns the key binding IDs that actions.go passes as string literals to
// registerKeyBindableAction and gurps.RegisterKeyBinding.
func keyBindingIDsInSource(c check.Checker) []string {
	file, err := parser.ParseFile(token.NewFileSet(), "actions.go", nil, 0)
	c.NoError(err, "actions.go must be parsable")
	var ids []string
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		switch fn := call.Fun.(type) {
		case *ast.Ident:
			if fn.Name != "registerKeyBindableAction" {
				return true
			}
		case *ast.SelectorExpr:
			if fn.Sel.Name != "RegisterKeyBinding" {
				return true
			}
		default:
			return true
		}
		if lit, ok2 := call.Args[0].(*ast.BasicLit); ok2 && lit.Kind == token.STRING {
			var id string
			if id, err = strconv.Unquote(lit.Value); err != nil {
				c.NoError(err, "key binding ID must be a valid string literal")
			} else {
				ids = append(ids, id)
			}
		}
		return true
	})
	return ids
}
