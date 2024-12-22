package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var _ BodySettingsOwner = &globalBodySettingsOwner{}

// BodySettingsOwner is the interface that a panel must implement to be able to display and edit Body settings.
type BodySettingsOwner interface {
	EntityPanel
	BodySettingsTitle() string
	BodySettings(forReset bool) *gurps.Body
	SetBodySettings(body *gurps.Body)
}

type globalBodySettingsOwner struct{}

func (g *globalBodySettingsOwner) AsPanel() *unison.Panel {
	return nil
}

func (g *globalBodySettingsOwner) Entity() *gurps.Entity {
	return nil
}

func (g *globalBodySettingsOwner) BodySettingsTitle() string {
	return i18n.Text("Default Body Type")
}

func (g *globalBodySettingsOwner) BodySettings(forReset bool) *gurps.Body {
	if forReset {
		return gurps.FactoryBody()
	}
	return gurps.GlobalSettings().Sheet.BodyType
}

func (g *globalBodySettingsOwner) SetBodySettings(body *gurps.Body) {
	gurps.GlobalSettings().Sheet.BodyType = body
}
