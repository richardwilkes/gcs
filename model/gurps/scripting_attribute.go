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
	"log/slog"

	"github.com/dop251/goja"
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

func newScriptAttribute(r *goja.Runtime, attr *Attribute) *goja.Object {
	m := make(map[string]func() goja.Value)
	m["id"] = func() goja.Value { return r.ToValue(attr.AttrID) }
	m["kind"] = func() goja.Value {
		if def := attr.AttributeDef(); def != nil {
			switch def.Kind() {
			case PrimaryAttrKind:
				return r.ToValue("primary")
			case SecondaryAttrKind:
				return r.ToValue("secondary")
			case PoolAttrKind:
				return r.ToValue("pool")
			}
		}
		return goja.Undefined()
	}
	m["name"] = func() goja.Value {
		if def := attr.AttributeDef(); def != nil {
			if def.Name == "" {
				return r.ToValue(def.FullName)
			}
			return r.ToValue(def.Name)
		}
		return goja.Undefined()
	}
	m["fullName"] = func() goja.Value {
		if def := attr.AttributeDef(); def != nil {
			return r.ToValue(def.ResolveFullName())
		}
		return goja.Undefined()
	}
	m["maximum"] = func() goja.Value {
		return maximumValueOfAttribute(r, attr)
	}
	m["current"] = func() goja.Value {
		if attr.Entity != nil {
			id := attr.AttrID + ".current"
			s := attr.Entity.ResolveVariable(id)
			v, err := fxp.FromString(s)
			if err != nil {
				slog.Error("failed to resolve attribute to number", "attr", id, "value", s)
				return goja.Undefined()
			}
			return r.ToValue(fxp.AsFloat[float64](v))
		}
		return goja.Undefined()
	}
	m["isDecimal"] = func() goja.Value {
		if def := attr.AttributeDef(); def != nil {
			return r.ToValue(def.AllowsDecimal())
		}
		return goja.Undefined()
	}
	m["valueOf"] = func() goja.Value {
		return r.ToValue(func(_ goja.FunctionCall) goja.Value {
			return maximumValueOfAttribute(r, attr)
		})
	}
	return r.NewDynamicObject(NewScriptObject(r, m))
}

func maximumValueOfAttribute(r *goja.Runtime, attr *Attribute) goja.Value {
	if attr.Entity != nil {
		s := attr.Entity.ResolveVariable(attr.AttrID)
		v, err := fxp.FromString(s)
		if err != nil {
			slog.Error("failed to resolve attribute to number", "attr", attr.AttrID, "value", s)
			return goja.Undefined()
		}
		return r.ToValue(fxp.AsFloat[float64](v))
	}
	return goja.Undefined()
}
