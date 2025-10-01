// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"reflect"
	"strings"

	"github.com/richardwilkes/toolbox/v2/xstrings"
)

type scriptNameMapper struct{}

func (n scriptNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string { //nolint:gocritic // API requires it
	return uncapitalizeScriptName(f.Name)
}

func (n scriptNameMapper) MethodName(_ reflect.Type, m reflect.Method) string { //nolint:gocritic // API requires it
	return uncapitalizeScriptName(m.Name)
}

func uncapitalizeScriptName(s string) string {
	if strings.EqualFold(s, "id") {
		return "id"
	}
	if strings.EqualFold(s, "parentid") {
		return "parentID"
	}
	return xstrings.FirstToLower(s)
}
