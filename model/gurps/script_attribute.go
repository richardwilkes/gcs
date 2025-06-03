package gurps

import (
	"log/slog"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

const unknown = "unknown"

type scriptAttribute struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Name      string  `json:"name"`
	FullName  string  `json:"fullName"`
	Maximum   float64 `json:"maximum"`
	Current   float64 `json:"current"`
	IsDecimal bool    `json:"isDecimal"`
}

func newScriptAttribute(attr *Attribute) *scriptAttribute {
	a := scriptAttribute{
		ID:       attr.ID(),
		Kind:     unknown,
		Name:     unknown,
		FullName: unknown,
	}
	if attr.Entity != nil {
		s := attr.Entity.ResolveVariable(attr.AttrID)
		v, err := fxp.FromString(s)
		if err != nil {
			slog.Error("failed to resolve attribute to number", "attr", attr.AttrID, "value", s)
		}
		a.Maximum = fxp.As[float64](v)
		id := attr.AttrID + ".current"
		s = attr.Entity.ResolveVariable(id)
		v, err = fxp.FromString(s)
		if err != nil {
			slog.Error("failed to resolve attribute to number", "attr", id, "value", s)
		}
		a.Current = fxp.As[float64](v)
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
		a.IsDecimal = def.AllowsDecimal()
	}
	return &a
}

func (a *scriptAttribute) ValueOf() float64 {
	return a.Maximum
}
