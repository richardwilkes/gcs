package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

const unknown = "unknown"

type scriptAttribute struct {
	ID           string  `json:"id"`
	Kind         string  `json:"kind"`
	Name         string  `json:"name"`
	FullName     string  `json:"fullName"`
	Maximum      fxp.Int `json:"maximum"`
	Current      fxp.Int `json:"current"`
	AllowDecimal bool    `json:"allowDecimal,omitempty"`
}

func newScriptAttribute(attr *Attribute) *scriptAttribute {
	a := scriptAttribute{
		ID:       attr.ID(),
		Kind:     unknown,
		Name:     unknown,
		FullName: unknown,
		Maximum:  attr.Maximum(),
		Current:  attr.Current(),
	}
	if def := attr.AttributeDef(); def != nil {
		switch def.Kind() {
		case PrimaryAttrKind:
			a.Kind = "primary"
		case SecondaryAttrKind:
			a.Kind = "secondary"
		case PoolAttrKind:
			a.Kind = "pool"
		}
		if def.Name == "" {
			a.Name = def.FullName
		} else {
			a.Name = def.Name
		}
		a.FullName = def.ResolveFullName()
		a.AllowDecimal = def.AllowsDecimal()
	}
	return &a
}

func (a *scriptAttribute) String() string {
	data, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
