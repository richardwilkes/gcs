package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

const unknown = "unknown"

type scriptAttribute struct {
	attr *Attribute
}

func (a *scriptAttribute) String() string {
	type attrJSON struct {
		ID           string  `json:"id"`
		Kind         string  `json:"kind"`
		Name         string  `json:"name"`
		FullName     string  `json:"fullName"`
		Maximum      fxp.Int `json:"maximum"`
		Current      fxp.Int `json:"current"`
		AllowDecimal bool    `json:"allowDecimal"`
	}
	data, err := json.Marshal(attrJSON{
		ID:           a.ID(),
		Kind:         a.Kind(),
		Name:         a.Name(),
		FullName:     a.FullName(),
		Maximum:      a.Maximum(),
		Current:      a.Current(),
		AllowDecimal: a.AllowDecimal(),
	})
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (a *scriptAttribute) ID() string {
	return a.attr.ID()
}

func (a *scriptAttribute) Kind() string {
	if def := a.attr.AttributeDef(); def != nil {
		switch def.Kind() {
		case PrimaryAttrKind:
			return "primary"
		case SecondaryAttrKind:
			return "secondary"
		case PoolAttrKind:
			return "pool"
		}
	}
	return unknown
}

func (a *scriptAttribute) Name() string {
	if def := a.attr.AttributeDef(); def != nil {
		if def.Name == "" {
			return def.FullName
		}
		return def.Name
	}
	return unknown
}

func (a *scriptAttribute) FullName() string {
	if def := a.attr.AttributeDef(); def != nil {
		return def.ResolveFullName()
	}
	return unknown
}

func (a *scriptAttribute) CombinedName() string {
	if def := a.attr.AttributeDef(); def != nil {
		return def.CombinedName()
	}
	return unknown
}

func (a *scriptAttribute) Current() fxp.Int {
	return a.attr.Current()
}

func (a *scriptAttribute) Maximum() fxp.Int {
	return a.attr.Maximum()
}

func (a *scriptAttribute) AllowDecimal() bool {
	if def := a.attr.AttributeDef(); def != nil {
		return def.AllowsDecimal()
	}
	return false
}
