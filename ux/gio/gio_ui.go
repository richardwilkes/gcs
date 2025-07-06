package gio

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/event" // For generic event type
	// Removed system import, will use app.DestroyEvent and app.FrameEvent
	"gioui.org/layout"
	"gioui.org/op"
	// "gioui.org/text" // Not used for now
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	// "gioui.org/font/gofont" // Not used if material.NewTheme() takes no args
)

var currentCharacter *CharacterSheet
var loadCharacterButton widget.Clickable

// StartGioUI initializes and runs the Gio UI.
func StartGioUI() {
	go func() {
		// Add recover for safety in the goroutine
		defer func() {
			if r := recover(); r != nil {
				log.Printf("!!!!!!!! Gio goroutine panicked: %v !!!!!!!!", r)
			}
			// os.Exit(1) // Optionally exit if Gio panics, or let app.Main handle exit
		}()

		// Log current working directory
		wd, errWd := os.Getwd()
		if errWd != nil {
			log.Printf("Error getting working directory: %v", errWd)
		} else {
			log.Printf("Gio UI Current Working Directory: %s", wd)
		}

		w := new(app.Window)
		w.Option(app.Title("GCS - Gio UI"))

		errLoop := loop(w) // Call the loop
		if errLoop != nil {
			log.Printf("Gio loop returned error: %v", errLoop)
		} else {
			log.Println("Gio loop exited cleanly.")
		}
		// The goroutine will exit here, app.Main() in the main goroutine will continue to manage the app lifecycle.
		// os.Exit should not be called from this goroutine directly unless it's a truly fatal, unrecoverable startup error.
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme()
	var ops op.Ops

	for {
		evt := w.Event()
		switch e := evt.(type) {
		case app.DestroyEvent:
			log.Println("DestroyEvent received, exiting loop.") // Log destroy event
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			if loadCharacterButton.Clicked(gtx) {
				log.Println("'Load Dai Blackthorn' button clicked.") // Log button click
				char, err := LoadCharacterSheet("dai_blackthorn.gcs")
				if err != nil {
					log.Printf("!!!!!!!! Error loading character sheet 'dai_blackthorn.gcs': %v !!!!!!!!", err)
				} else {
					currentCharacter = char
					log.Println("Successfully loaded Dai Blackthorn from dai_blackthorn.gcs")
				}
			}

			layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Button(th, &loadCharacterButton, "Load Dai Blackthorn").Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if currentCharacter == nil {
						// Using simple material.Label, assuming Body1/H6 might not be stable/exist in this version
						return material.Label(th, unit.Sp(16), "No character loaded.").Layout(gtx)
					}
					return layoutCharacterSheet(th, gtx, currentCharacter)
				}),
			)
			e.Frame(gtx.Ops)
		case app.ConfigEvent: // Handle config changes if necessary
			log.Printf("Config event: %v", e)
		case event.Event: // Catch-all for other/unhandled events
			// log.Printf("Generic event: %T %v", e, e)
		}
	}
}

func layoutCharacterSheet(th *material.Theme, gtx layout.Context, char *CharacterSheet) layout.Dimensions {
	// Revert to simpler label styling, avoiding H6/Body1/Body2 if they are version-specific
	hSubtitle := func(th *material.Theme, txtStr string) material.LabelStyle {
		lbl := material.Label(th, unit.Sp(20), txtStr)
		// Font weight styling removed to ensure compilation with potentially older API
		return lbl
	}
	body1 := func(th *material.Theme, txtStr string) material.LabelStyle {
		return material.Label(th, unit.Sp(16), txtStr)
	}
	body2 := func(th *material.Theme, txtStr string) material.LabelStyle {
		return material.Label(th, unit.Sp(14), txtStr)
	}

	return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceEnd, Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, hSubtitle(th, fmt.Sprintf("Name: %s", char.Profile.Name)).Layout),
				layout.Flexed(1, hSubtitle(th, fmt.Sprintf("Player: %s", char.Profile.PlayerName)).Layout),
				layout.Flexed(1, hSubtitle(th, fmt.Sprintf("Points: %d", char.TotalPoints)).Layout),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(body1(th, "Profile").Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(body2(th, fmt.Sprintf("Age: %s, Gender: %s", char.Profile.Age, char.Profile.Gender)).Layout),
					layout.Rigid(body2(th, fmt.Sprintf("Height: %s, Weight: %s", char.Profile.Height, char.Profile.Weight)).Layout),
					layout.Rigid(body2(th, fmt.Sprintf("Eyes: %s, Hair: %s, Skin: %s", char.Profile.Eyes, char.Profile.Hair, char.Profile.Skin)).Layout),
				)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(body1(th, "Attributes").Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layoutAttributesColumn(th, gtx, char, true, body2)
					}),
					layout.Flexed(0.5, func(gtx layout.Context) layout.Dimensions {
						return layoutAttributesColumn(th, gtx, char, false, body2)
					}),
				)
			})
		}),
	)
}

func layoutAttributesColumn(th *material.Theme, gtx layout.Context, char *CharacterSheet, primary bool, textStyle func(*material.Theme, string) material.LabelStyle) layout.Dimensions {
	var flexChildren []layout.FlexChild
	primaryAttrs := map[string]bool{"st": true, "dx": true, "iq": true, "ht": true}

	for _, attrVal := range char.Attributes {
		isPrimary := primaryAttrs[attrVal.AttrID]
		if (primary && isPrimary) || (!primary && !isPrimary) {
			def := findAttributeDef(char.Settings.Attributes, attrVal.AttrID)
			if def != nil {
				labelStr := fmt.Sprintf("%s (%s): %d", def.Name, def.ID, attrVal.Calc.Value)
				if def.Type == "pool" {
					labelStr += fmt.Sprintf(" (Current: %d)", attrVal.Calc.Current)
				}
				flexChildren = append(flexChildren, layout.Rigid(textStyle(th, labelStr).Layout))
			}
		}
	}
	if len(flexChildren) == 0 {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, flexChildren...)
}

func findAttributeDef(defs []AttributeDef, id string) *AttributeDef {
	for i := range defs {
		if defs[i].ID == id {
			return &defs[i]
		}
	}
	return nil
}
