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
	"math"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const (
	// appUpdatePulsePeriod is how long one full fade out and back of the update button takes.
	appUpdatePulsePeriod = 1500 * time.Millisecond
	// appUpdatePulseTick is the interval between adjustments of the update button's color.
	appUpdatePulseTick = time.Second / 20
	// appUpdatePulseDuration is how long the update button pulses after a release appears before settling into its
	// resting color. Repainting the toolbar twenty times a second for the rest of a session in which the user has
	// decided to update later would keep a laptop awake for nothing, and two minutes is long enough to be noticed.
	appUpdatePulseDuration = 2 * time.Minute
)

// SyncAppUpdateButton brings the Library Explorer's "Software Update Available" button into line with the current
// update state and settings. Must be called on the UI thread; the update checker posts it there via unison.InvokeTask.
// Does nothing if the Library Explorer hasn't been built yet, since it syncs itself as it is created.
func SyncAppUpdateButton() {
	if Workspace.Navigator != nil && Workspace.Navigator.appUpdateButton != nil {
		Workspace.Navigator.syncAppUpdateButton()
	}
}

// newAppUpdateButton creates the filled, attention-getting button that is added to the toolbar when a newer version of
// the application is available. It isn't a child of anything until then: a hidden panel would still take up space in
// the toolbar's flow layout, and even one reporting no size would still have the layout's spacing added after it.
func newAppUpdateButton(click func()) *unison.Button {
	button := unison.NewSVGButton(svg.Download)
	// Unlike the other toolbar buttons, this one is a filled chip in the warning colors, so that it stands out.
	button.HideBase = false
	button.BackgroundInk = unison.ThemeWarning
	button.OnBackgroundInk = unison.ThemeOnWarning
	button.SetTitle(i18n.Text("Software Update Available"))
	button.ClickCallback = click
	button.SetLayoutData(align.Middle) // The same alignment the toolbar gives the rest of its row
	return button
}

// showAppUpdateButton returns true if the update button should be visible. With the setting at Never, the button is
// never shown, even for an update that is already known from an earlier check or from an explicit check made through
// the Help menu. The caller reads the first release, so an empty list counts as no update, whether or not it is nil.
func showAppUpdateButton(releases []gurps.Release, option updatecheck.Option) bool {
	return len(releases) != 0 && option != updatecheck.Never
}

// syncAppUpdateButton adds the update button to the toolbar or removes it, updates its tooltip, and starts or stops its
// pulse.
func (n *Navigator) syncAppUpdateButton() {
	title, releases, _ := AppUpdateResult()
	if showAppUpdateButton(releases, gurps.GlobalSettings().General.AppUpdateCheck) {
		n.appUpdateButton.Tooltip = newWrappedTooltip(title)
		if n.appUpdateButton.Parent() == nil {
			n.buttonRow.AddChild(n.appUpdateButton)
		}
		n.appUpdatePulse.start(releases[0].Version)
	} else {
		n.appUpdatePulse.stop()
		n.appUpdateButton.RemoveFromParent()
	}
	// Adding or removing the button changes the row's size, so everything above it has to lay out again: unison only
	// lays out the panels whose own NeedsLayout flag is set, and laying out the row alone would leave its ancestors
	// with the frames they had before.
	n.buttonRow.MarkForLayoutRecursivelyUpward()
	n.buttonRow.MarkForRedraw()
}

// applyAppUpdatePulse tints the update button for the given point in the pulse cycle. At rest the button is given the
// theme's own ink rather than a snapshot of its color, so that a change of theme reaches it. The blend is limited to
// half of the way to the surrounding toolbar color so that the ThemeOnWarning text and icon remain readable in both the
// light and dark themes.
func (n *Navigator) applyAppUpdatePulse(pct float32) {
	if pct <= 0 {
		n.appUpdateButton.BackgroundInk = unison.ThemeWarning
	} else {
		n.appUpdateButton.BackgroundInk = unison.ThemeWarning.GetColor().Blend(unison.ThemeAboveSurface.GetColor(),
			pct*0.5)
	}
	n.appUpdateButton.MarkForRedraw()
}

// appUpdatePulse repeatedly hands out the current point in a pulse cycle, for appUpdatePulseDuration after each new
// release, and then settles. A scheduled tick can't be canceled, so each one carries the generation it was scheduled
// in; starting or stopping bumps the generation, which retires anything already pending. All of its methods must be
// called on the UI thread, which is also where unison.InvokeTaskAfter delivers the ticks.
type appUpdatePulse struct {
	apply     func(pct float32)                   // What to do with each new point in the cycle
	schedule  func(f func(), after time.Duration) // nil means unison.InvokeTaskAfter; tests inject a recorder
	now       func() time.Time                    // nil means time.Now; tests inject a clock
	startedAt time.Time
	version   string // The release the pulse last ran for
	gen       int
	running   bool
}

// start begins pulsing for the given release. The pulse runs for appUpdatePulseDuration and then settles, and it
// doesn't start again for a release it has already pulsed for: the syncs that follow every check must neither restart
// the cycle nor wake a pulse that has settled. Only a release the button hasn't announced before sets it going again,
// or the button being shown again after having been hidden (see stop).
func (p *appUpdatePulse) start(version string) {
	if p.version == version {
		return
	}
	p.version = version
	p.gen++
	p.running = true
	p.startedAt = p.clock()
	p.scheduleTick(p.gen)
}

// stop ends the pulsing, retires any tick already pending, and forgets the release, so that the button pulses again if
// it is shown again for the same one.
func (p *appUpdatePulse) stop() {
	p.settle()
	p.version = ""
}

// settle ends the pulsing, retires any tick already pending, and puts the button back at rest.
func (p *appUpdatePulse) settle() {
	p.gen++
	if !p.running {
		return
	}
	p.running = false
	if p.apply != nil {
		p.apply(0)
	}
}

func (p *appUpdatePulse) scheduleTick(gen int) {
	schedule := p.schedule
	if schedule == nil {
		schedule = unison.InvokeTaskAfter
	}
	schedule(func() { p.tick(gen) }, appUpdatePulseTick)
}

func (p *appUpdatePulse) clock() time.Time {
	if p.now != nil {
		return p.now()
	}
	return time.Now()
}

// tick applies the current point in the cycle and sets up the next one, or settles the pulse once it has run for long
// enough. Ticks left over from a stopped or restarted pulse are ignored.
func (p *appUpdatePulse) tick(gen int) {
	if gen != p.gen || !p.running {
		return
	}
	elapsed := p.clock().Sub(p.startedAt)
	if elapsed >= appUpdatePulseDuration {
		p.settle()
		return
	}
	if p.apply != nil {
		p.apply(pulseFraction(elapsed, appUpdatePulsePeriod))
	}
	p.scheduleTick(gen)
}

// pulseFraction returns where in a pulse cycle of the given period the given elapsed time falls, as a value in the
// range [0,1] that rises from 0 to 1 over the first half of the period and falls back to 0 over the second half. The
// triangle wave is eased with a half-cosine so that the pulse doesn't visibly jerk at the ends of its travel.
func pulseFraction(elapsed, period time.Duration) float32 {
	if period <= 0 {
		return 0
	}
	pos := elapsed % period
	if pos < 0 {
		pos += period
	}
	t := 2 * float64(pos) / float64(period)
	if t > 1 {
		t = 2 - t
	}
	return min(max(float32((1-math.Cos(math.Pi*t))/2), 0), 1)
}
