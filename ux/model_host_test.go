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
	"bytes"
	"encoding/json/jsontext"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/canvas/codecs"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/thememode"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/colors"
	"github.com/richardwilkes/gcs/v5/ux/fonts"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	unthememode "github.com/richardwilkes/unison/enums/thememode"
)

// TestModelDefaultsMatchUnison verifies the values the model has to hold copies of, since it can't ask unison for
// them: the cursor size limits, and the scroll wheel multiplier it falls back on when there is no host.
func TestModelDefaultsMatchUnison(t *testing.T) {
	c := check.New(t)
	c.Equal(int(unison.DefaultCursorSize().Width), gurps.CursorSizeDef)
	c.Equal(unison.MinCursorSize, gurps.CursorSizeMin)
	c.Equal(unison.MaxCursorSize, gurps.CursorSizeMax)

	c.Equal(fxp.FromFloat(unison.MouseWheelMultiplier), gurps.NewGeneralSettings().ScrollWheelMultiplier,
		"the host supplies unison's multiplier")
	c.Equal(unison.MouseWheelMultiplier, gurps.NoHost{}.DefaultScrollWheelMultiplier(),
		"the model's own fallback is the same")
}

// TestGeneralSettingsReachUnison verifies that validating the general settings hands those unison keeps a copy of to
// it.
func TestGeneralSettingsReachUnison(t *testing.T) {
	c := check.New(t)
	savedSize := unison.CursorSize()
	savedFocus := unison.FocusForReading()
	savedDelay := unison.DefaultTooltipTheme.Delay
	savedDismissal := unison.DefaultTooltipTheme.Dismissal
	t.Cleanup(func() {
		unison.SetCursorSize(savedSize)
		unison.SetFocusForReading(savedFocus)
		unison.DefaultTooltipTheme.Delay = savedDelay
		unison.DefaultTooltipTheme.Dismissal = savedDismissal
	})

	s := gurps.NewGeneralSettings()
	s.CursorSize = gurps.CursorSizeMin
	s.FocusForReading = !savedFocus
	s.TooltipDelay = fxp.Two
	s.TooltipDismissal = fxp.Ten
	s.EnsureValidity()
	c.Equal(geom.NewSize(gurps.CursorSizeMin, gurps.CursorSizeMin), unison.CursorSize())
	c.Equal(!savedFocus, unison.FocusForReading())
	c.Equal(fxp.SecondsToDuration(fxp.Two), unison.DefaultTooltipTheme.Delay)
	c.Equal(fxp.SecondsToDuration(fxp.Ten), unison.DefaultTooltipTheme.Dismissal)
}

// TestDisplayPPI verifies that a display that is missing, which happens on some Linux configurations when no monitor is
// enumerated, or that reports no content scale yields 0 rather than a panic or a division by zero, and that the model
// then falls back on a usable PPI.
func TestDisplayPPI(t *testing.T) {
	c := check.New(t)
	c.Equal(0, displayPPI(nil))
	c.Equal(0, displayPPI(&unison.Display{PPI: 216, Scale: geom.Point{}}))
	c.Equal(0, displayPPI(&unison.Display{PPI: 0, Scale: geom.NewPoint(2, 2)}))
	c.Equal(108, displayPPI(&unison.Display{PPI: 216, Scale: geom.NewPoint(2, 2)}))
	c.Equal(216, displayPPI(&unison.Display{PPI: 216, Scale: geom.NewPoint(1, 1)}))
	c.True((&gurps.GeneralSettings{}).MonitorPPI() > 0)
}

// TestModelHostCarriesThemeInSettings verifies that the live theme colors and fonts are stored into the settings when
// they are saved and applied from them when they are loaded.
func TestModelHostCarriesThemeInSettings(t *testing.T) {
	c := check.New(t)
	h := modelHost{}
	g := gurps.GlobalSettings()
	swapForTest(t, &g.Colors, nil)
	swapForTest(t, &g.Fonts, nil)
	original := *colors.Header
	font := fonts.CurrentFonts()[0]
	originalFont := font.Font.Font
	t.Cleanup(func() {
		*colors.Header = original
		font.Font.Font = originalFont
		unison.ThemeChanged()
	})

	custom := unison.ThemeColor{Light: unison.RGB(1, 2, 3), Dark: unison.RGB(4, 5, 6)}
	*colors.Header = custom
	customFont := originalFont.Face().Font(originalFont.Size() + 3)
	font.Font.Font = customFont
	css, ok := h.ThemeColor("header")
	c.True(ok)
	c.Equal(custom.GetColor().String(), css)
	_, ok = h.ThemeColor("bogus")
	c.False(ok)

	h.SettingsSaving(g)
	var stored map[string]unison.ThemeColor
	c.NoError(jio.Unmarshal(g.Colors, &stored))
	c.Equal(custom, stored["header"], "the live color is stored into the settings")
	c.Equal(len(colors.Current()), len(stored))
	var storedFonts map[string]unison.FontDescriptor
	c.NoError(jio.Unmarshal(g.Fonts, &storedFonts))
	c.Equal(customFont.Descriptor(), storedFonts[font.ID], "as is the live font")
	c.Equal(len(fonts.CurrentFonts()), len(storedFonts))

	*colors.Header = original
	font.Font.Font = originalFont
	h.SettingsLoaded(g)
	c.Equal(custom, *colors.Header, "the stored color is applied to the live theme")
	c.Equal(customFont.Descriptor(), font.Font.Descriptor(), "as is the stored font")
}

// TestModelHostKeepsDecodableThemeValues verifies that a theme value the settings carry that can't be decoded costs
// only itself: the others are still applied, and it is the only one that comes from the factory instead.
func TestModelHostKeepsDecodableThemeValues(t *testing.T) {
	c := check.New(t)
	g := gurps.GlobalSettings()
	swapForTest(t, &g.Colors, nil)
	swapForTest(t, &g.Fonts, nil)
	original := *colors.Header
	originalSuccess := *colors.Success
	t.Cleanup(func() {
		*colors.Header = original
		*colors.Success = originalSuccess
		unison.ThemeChanged()
	})

	custom := unison.ThemeColor{Light: unison.RGB(1, 2, 3), Dark: unison.RGB(4, 5, 6)}
	customJSON, err := jio.Marshal(custom)
	c.NoError(err)
	*colors.Success = custom
	g.Colors = jsontext.Value(`{"header":` + string(customJSON) + `,"success":5}`)
	modelHost{}.SettingsLoaded(g)
	c.Equal(custom, *colors.Header, "the value that could be decoded is applied")
	factorySuccess := slices.IndexFunc(colors.Factory(), func(one *colors.ThemedColor) bool { return one.ID == "success" })
	c.True(factorySuccess >= 0)
	c.Equal(colors.Factory()[factorySuccess].Color.GetColor(), colors.Success.GetColor(),
		"the one that couldn't be is reset to the factory value")
}

// TestUnisonThemeMode verifies the mapping from the model's theme modes to unison's.
func TestUnisonThemeMode(t *testing.T) {
	c := check.New(t)
	c.Equal(len(unthememode.All), len(thememode.Modes))
	for _, mode := range thememode.Modes {
		c.Equal(mode.Key(), unisonThemeMode(mode).Key())
	}
}

// TestThemeFileConvertersAreRegistered verifies that the theme color and font files, whose content only the user
// interface understands, are brought up to date along with the files the model converts itself.
func TestThemeFileConvertersAreRegistered(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	colorsPath := filepath.Join(dir, "theme"+gurps.ColorSettingsExt)
	c.NoError(os.WriteFile(colorsPath, fmt.Appendf(nil, `{"version":%d,"colors":{}}`, jio.CurrentDataVersion), 0o600))
	fontsPath := filepath.Join(dir, "theme"+gurps.FontSettingsExt)
	c.NoError(os.WriteFile(fontsPath, fmt.Appendf(nil, `{"version":%d,"fonts":{}}`, jio.CurrentDataVersion-1), 0o600))

	c.NoError(gurps.Convert(dir))

	data, err := os.ReadFile(colorsPath)
	c.NoError(err)
	c.Contains(string(data), `"header"`, "the color file is rewritten with every color in it")
	data, err = os.ReadFile(fontsPath)
	c.NoError(err)
	c.Contains(string(data), `"system"`, "the font file is rewritten with every font in it")
	c.Contains(string(data), fmt.Sprintf(`"version": %d`, jio.CurrentDataVersion), "in the current format")
}

// TestPortraitImage verifies that a portrait whose bytes the current build cannot decode is neither discarded, since a
// different build may be able to decode it, nor decoded over and over, and that replacing the data lets a valid image
// be loaded afterwards.
func TestPortraitImage(t *testing.T) {
	c := check.New(t)
	data := []byte("this is not a valid image")
	var p gurps.Profile
	p.PortraitData = data
	c.Nil(portraitImage(&p))
	c.Equal(data, p.PortraitData)
	cached := p.PortraitCache
	c.NotNil(cached, "the failure is remembered")

	c.Nil(portraitImage(&p))
	c.True(cached == p.PortraitCache, "a second call must not decode again")
	c.Equal(data, p.PortraitData)

	codecs.Register() // unison installs the image decoders as the app starts, which a plain test never does.
	var buffer bytes.Buffer
	c.NoError(png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	p.SetPortraitData(buffer.Bytes())
	img := portraitImage(&p)
	c.NotNil(img)
	c.True(img == portraitImage(&p), "the decoded image is reused")

	p.SetPortraitData(nil)
	c.Nil(portraitImage(&p))
}

// TestDockStateEncoding verifies that a dock state survives being carried by the settings as raw JSON, and that the
// absence of one is preserved.
func TestDockStateEncoding(t *testing.T) {
	c := check.New(t)
	c.Nil(decodeDockState(nil))
	c.Nil(decodeDockState(jsontext.Value("null")))
	c.Equal(0, len(encodeDockState(nil)))

	state := &unison.DockState{
		Type:     unison.ContainerType,
		Children: []*unison.DockState{{Type: unison.DockableType, Key: NavigatorDockKey}},
	}
	decoded := decodeDockState(encodeDockState(state))
	c.NotNil(decoded)
	c.Equal(unison.ContainerType, decoded.Type)
	c.Equal([]string{NavigatorDockKey}, dockStateKeys(decoded))
}

// TestHAlignFor verifies the mapping from the model's cell alignments to unison's.
func TestHAlignFor(t *testing.T) {
	c := check.New(t)
	c.Equal(align.Start, hAlignFor(cell.AlignStart))
	c.Equal(align.Middle, hAlignFor(cell.AlignMiddle))
	c.Equal(align.End, hAlignFor(cell.AlignEnd))
}

// TestApplyKeyBindingsStoresCanonicalText verifies that a binding spelled differently than unison would spell it is
// recognized for what it is, rather than being kept as a change from the factory default it is equal to.
func TestApplyKeyBindingsStoresCanonicalText(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	var id, key, respelled string
	for _, one := range currentKeyBindings() {
		key = one.KeyBinding.Key()
		if respelled = strings.ToUpper(key); respelled == key {
			respelled = strings.ToLower(key)
		}
		if respelled != key && one.KeyBinding == keyBindingEntries[one.ID].KeyBinding {
			id = one.ID
			break
		}
	}
	c.NotEqual("", id, "there must be a key binding at its factory default whose text has letters in it")
	t.Cleanup(func() { applyKeyBindings(&gurps.GlobalSettings().KeyBindings) })

	var untouched gurps.KeyBindings
	applyKeyBindings(&untouched)
	c.True(untouched.IsZero(), "applying the factory defaults must not record any of them as a change")
	for _, entry := range keyBindingEntries {
		c.Equal(entry.KeyBinding, entry.Action.KeyBinding,
			"%s: the binding must survive the trip through its text, not just the text", entry.ID)
	}

	var b gurps.KeyBindings
	b.Set(id, respelled)
	c.False(b.IsZero(), "the model can only compare the text")
	applyKeyBindings(&b)
	c.Equal(key, b.Current(id))
	c.True(b.IsZero(), "once respelled, the binding is seen to be the factory default")
	c.Equal(keyBindingEntries[id].KeyBinding, keyBindingEntries[id].Action.KeyBinding)
}
