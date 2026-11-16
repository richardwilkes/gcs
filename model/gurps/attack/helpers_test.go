// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack_test

import "github.com/richardwilkes/toolbox/v2/check"

// parseCase pairs an input with the string a weapon field parsed from it is expected to render as.
type parseCase struct {
	input    string
	expected string
}

// checkFieldParsing verifies parse, which parses one kind of weapon field and renders it back to a string: each of the
// same strings must come back unchanged, and each adjusted input must come back as its expected string.
func checkFieldParsing(c check.Checker, parse func(string) string, same []string, adjusted []parseCase) {
	c.Helper()
	for i, s := range same {
		c.Equal(s, parse(s), "same[%d] %q", i, s)
	}
	for i, one := range adjusted {
		c.Equal(one.expected, parse(one.input), "adjusted[%d] %q", i, one.input)
	}
}

// validateCase pairs a weapon field with the value Validate is expected to turn it into.
type validateCase[T any] struct {
	name     string
	in       T
	expected T
}

// checkValidate verifies that Validate turns each case's value into its expected value.
func checkValidate[T any, P interface {
	*T
	Validate()
}](c check.Checker, cases []validateCase[T]) {
	c.Helper()
	for _, one := range cases {
		P(&one.in).Validate()
		c.Equal(one.expected, one.in, one.name)
	}
}
