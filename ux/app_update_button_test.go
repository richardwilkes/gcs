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
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
)

// pulseTolerance is how far a computed pulse fraction may sit from the expected one, which allows for the float32
// rounding in the cosine easing.
const pulseTolerance = 0.0001

// TestPulseFractionEndpoints verifies the shape of one pulse cycle: it starts dark, reaches full brightness halfway
// through, and is back where it started at the end, so that repeating it produces a smooth throb rather than a sawtooth
// that snaps back.
func TestPulseFractionEndpoints(t *testing.T) {
	c := check.New(t)
	period := appUpdatePulsePeriod
	for _, one := range []struct {
		name     string
		elapsed  time.Duration
		expected float32
	}{
		{name: "start", elapsed: 0, expected: 0},
		{name: "quarter", elapsed: period / 4, expected: 0.5},
		{name: "half", elapsed: period / 2, expected: 1},
		{name: "three quarters", elapsed: 3 * period / 4, expected: 0.5},
		{name: "full", elapsed: period, expected: 0},
		{name: "second cycle half", elapsed: 3 * period / 2, expected: 1},
	} {
		actual := pulseFraction(one.elapsed, period)
		c.True(xmath.Abs(actual-one.expected) <= pulseTolerance,
			"at %s expected %v, got %v", one.name, one.expected, actual)
	}
}

// TestPulseFractionStaysInRange verifies that a sweep across several cycles never produces a value outside [0,1], since
// the fraction is fed to a color blend that would otherwise be clamped and stall the pulse at an end of its travel.
func TestPulseFractionStaysInRange(t *testing.T) {
	c := check.New(t)
	period := appUpdatePulsePeriod
	for i := range 1000 {
		elapsed := time.Duration(i) * period / 100
		pct := pulseFraction(elapsed, period)
		c.True(pct >= 0 && pct <= 1, "expected a fraction within [0,1] at %v, got %v", elapsed, pct)
	}
}

// TestPulseFractionWithoutPeriod verifies that a non-positive period yields a stable 0 rather than dividing by zero.
func TestPulseFractionWithoutPeriod(t *testing.T) {
	c := check.New(t)
	c.Equal(float32(0), pulseFraction(time.Second, 0), "a zero period must not pulse")
	c.Equal(float32(0), pulseFraction(time.Second, -time.Second), "a negative period must not pulse")
}

// TestShowAppUpdateButton verifies when the "Software Update Available" button is revealed. Most importantly, with the
// setting at Never it is never shown, even for an update already known from an earlier or manual check.
func TestShowAppUpdateButton(t *testing.T) {
	c := check.New(t)
	releases := []gurps.Release{{Version: "5.99.0"}}
	for _, one := range []struct {
		name     string
		releases []gurps.Release
		option   updatecheck.Option
		expected bool
	}{
		{name: "no update known", releases: nil, option: updatecheck.AtLaunch, expected: false},
		{name: "no update known, never", releases: nil, option: updatecheck.Never, expected: false},
		// The caller reads the first release, so an empty list must count as no update even though it isn't nil.
		{name: "empty update list", releases: []gurps.Release{}, option: updatecheck.AtLaunch, expected: false},
		{name: "update known, at launch", releases: releases, option: updatecheck.AtLaunch, expected: true},
		{name: "update known, hourly", releases: releases, option: updatecheck.Hourly, expected: true},
		{name: "update known, daily", releases: releases, option: updatecheck.Daily, expected: true},
		{name: "update known, never", releases: releases, option: updatecheck.Never, expected: false},
	} {
		c.Equal(one.expected, showAppUpdateButton(one.releases, one.option), "%s", one.name)
	}
}

// pulseRecorder stands in for unison.InvokeTaskAfter, capturing the ticks a pulse schedules so that the tests can drive
// them by hand without a UI thread.
type pulseRecorder struct {
	pending []func()
	delays  []time.Duration
}

func (r *pulseRecorder) schedule(f func(), after time.Duration) {
	r.pending = append(r.pending, f)
	r.delays = append(r.delays, after)
}

// fireAll runs the ticks captured so far, allowing them to schedule new ones.
func (r *pulseRecorder) fireAll() {
	pending := r.pending
	r.pending = nil
	for _, f := range pending {
		f()
	}
}

// TestAppUpdatePulseScheduling verifies that the pulse keeps exactly one tick in flight while running and none once
// stopped. A scheduled tick can't be canceled, so a leftover tick arriving after a stop must do nothing: without the
// generation check it would resume the pulse and repaint a button that is no longer visible, and a second start()
// while running would leave two self-perpetuating tick chains behind.
func TestAppUpdatePulseScheduling(t *testing.T) {
	c := check.New(t)
	var recorder pulseRecorder
	applied := make([]float32, 0, 4)
	p := appUpdatePulse{
		apply:    func(pct float32) { applied = append(applied, pct) },
		schedule: recorder.schedule,
	}

	p.start("99.0.0")
	c.Equal(1, len(recorder.pending), "start must schedule one tick")
	c.Equal(appUpdatePulseTick, recorder.delays[0], "the tick must be scheduled for the pulse interval")

	p.start("99.0.0")
	c.Equal(1, len(recorder.pending), "starting an already running pulse must not schedule a second tick")

	recorder.fireAll()
	c.Equal(1, len(applied), "a tick must apply the current point in the cycle")
	c.Equal(1, len(recorder.pending), "a tick must schedule the next one")

	p.stop()
	c.Equal(2, len(applied), "stopping must put the button back at rest")
	c.Equal(float32(0), applied[1])
	recorder.fireAll()
	c.Equal(2, len(applied), "a tick left over from a stopped pulse must not apply anything")
	c.Equal(0, len(recorder.pending), "a stopped pulse must not schedule further ticks")

	p.start("99.0.0")
	c.Equal(1, len(recorder.pending), "a stopped pulse must be able to start again for the same release")
}

// TestAppUpdatePulseSettlesAndRestartsOnlyForANewRelease verifies that the pulse gives up after appUpdatePulseDuration,
// leaving the button at rest, and that the syncs which follow every later check don't set it going again for the
// release it has already announced. Without the limit the toolbar was repainted twenty times a second for the rest of
// any session in which the user chose to update later; without the memory of the release, the next hourly check would
// have started the two minutes over.
func TestAppUpdatePulseSettlesAndRestartsOnlyForANewRelease(t *testing.T) {
	c := check.New(t)
	var recorder pulseRecorder
	applied := make([]float32, 0, 4)
	clock := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.UTC)
	p := appUpdatePulse{
		apply:    func(pct float32) { applied = append(applied, pct) },
		schedule: recorder.schedule,
		now:      func() time.Time { return clock },
	}

	p.start("99.0.0")
	clock = clock.Add(appUpdatePulseDuration - appUpdatePulseTick)
	recorder.fireAll()
	c.Equal(1, len(applied), "a tick before the duration is up must apply a point in the cycle")
	c.Equal(1, len(recorder.pending), "...and schedule the next one")

	clock = clock.Add(appUpdatePulseTick)
	recorder.fireAll()
	c.Equal(2, len(applied), "the tick that finds the duration up must put the button at rest")
	c.Equal(float32(0), applied[1])
	c.Equal(0, len(recorder.pending), "a settled pulse must not schedule further ticks")
	c.False(p.running)

	p.start("99.0.0")
	c.Equal(0, len(recorder.pending), "a settled pulse must not start again for the release it settled on")

	p.start("99.1.0")
	c.Equal(1, len(recorder.pending), "a release that hasn't been announced must start the pulse again")
	c.True(p.running)

	// Hiding the button and showing it again is a fresh announcement, so the same release pulses again.
	p.stop()
	recorder.fireAll()
	c.Equal(0, len(recorder.pending))
	p.start("99.1.0")
	c.Equal(1, len(recorder.pending), "a pulse that was stopped must start again for the same release")
}

// seedAppUpdate puts the given releases into the shared update state for the duration of the test, with nil meaning
// that nothing is known, and restores what was there afterwards.
func seedAppUpdate(t *testing.T, releases []gurps.Release) {
	t.Helper()
	appUpdate.lock.Lock()
	savedResult, savedReleases, savedUpdating := appUpdate.result, appUpdate.releases, appUpdate.updating
	appUpdate.lock.Unlock()
	t.Cleanup(func() {
		appUpdate.lock.Lock()
		appUpdate.result, appUpdate.releases, appUpdate.updating = savedResult, savedReleases, savedUpdating
		appUpdate.lock.Unlock()
	})
	if releases != nil {
		appUpdate.SetReleases(releases)
		return
	}
	appUpdate.lock.Lock()
	appUpdate.result, appUpdate.releases, appUpdate.updating = "", nil, false
	appUpdate.lock.Unlock()
}

// TestSyncAppUpdateButtonAddsAndRemovesIt verifies that the update button is a child of the toolbar row only while
// there is an update to announce, and is added just once however many syncs say so. It used to sit in the row hidden
// and reporting no size, but the row's flow layout adds its spacing after every child regardless of size, so even a
// hidden button left a gap.
func TestSyncAppUpdateButtonAddsAndRemovesIt(t *testing.T) {
	c := check.New(t)
	gs, _ := prepareUpdateCheckSettings(t)
	var recorder pulseRecorder
	n := &Navigator{
		buttonRow:       unison.NewPanel(),
		appUpdateButton: newAppUpdateButton(nil),
	}
	n.appUpdatePulse.apply = n.applyAppUpdatePulse
	n.appUpdatePulse.schedule = recorder.schedule

	seedAppUpdate(t, nil)
	n.syncAppUpdateButton()
	c.Equal(0, len(n.buttonRow.Children()), "with no update known, the button must not be in the row")
	c.Equal(0, len(recorder.pending), "with no update known, nothing must pulse")

	seedAppUpdate(t, pendingAppRelease)
	n.syncAppUpdateButton()
	c.Equal(1, len(n.buttonRow.Children()), "an update must add the button to the row")
	c.True(n.buttonRow.Children()[0].Is(n.appUpdateButton))
	c.Equal(1, len(recorder.pending), "an update must start the pulse")

	n.syncAppUpdateButton()
	c.Equal(1, len(n.buttonRow.Children()), "a second sync must not add the button again")
	c.Equal(1, len(recorder.pending), "a second sync must not restart the pulse")

	gs.AppUpdateCheck = updatecheck.Never
	n.syncAppUpdateButton()
	c.Equal(0, len(n.buttonRow.Children()), "turning the checks off must take the button out of the row")
	c.Equal((*unison.Panel)(nil), n.appUpdateButton.Parent())
	recorder.fireAll()
	c.Equal(0, len(recorder.pending), "turning the checks off must stop the pulse")
}
