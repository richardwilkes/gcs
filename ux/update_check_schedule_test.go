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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// recordedTick is one request the periodicCheck made of its scheduler.
type recordedTick struct {
	fire  func()
	after time.Duration
}

// periodicCheckHarness drives a periodicCheck without unison and without a clock. The scheduler records what it is
// asked to run rather than running it, both because the tests need to fire ticks at moments of their choosing and
// because a scheduler that called back immediately would recurse: a tick reschedules before it runs the work.
type periodicCheckHarness struct {
	check  periodicCheck
	option updatecheck.Option
	ticks  []recordedTick
	runs   int
}

// newPeriodicCheckHarness returns a harness whose check starts out set to the given option.
func newPeriodicCheckHarness(option updatecheck.Option) *periodicCheckHarness {
	h := &periodicCheckHarness{option: option}
	h.check = periodicCheck{
		frequency: func() updatecheck.Option { return h.option },
		run:       func() { h.runs++ },
		schedule: func(f func(), after time.Duration) {
			h.ticks = append(h.ticks, recordedTick{fire: f, after: after})
		},
	}
	return h
}

// The options that don't repeat -- Never, and the default of checking only at launch -- must leave no tick pending.
// Anything else would keep waking the application up to do work the user asked it not to do.
func TestPeriodicCheckSchedulesNothingWithoutAnInterval(t *testing.T) {
	c := check.New(t)
	for _, option := range []updatecheck.Option{updatecheck.Never, updatecheck.AtLaunch} {
		h := newPeriodicCheckHarness(option)
		h.check.reschedule()
		c.Equal(0, len(h.ticks), "%s must not schedule a repeating check", option.Key())
		c.False(h.check.pending, "%s must leave nothing pending", option.Key())
	}
}

func TestPeriodicCheckSchedulesAtTheRequestedInterval(t *testing.T) {
	c := check.New(t)
	for option, want := range map[updatecheck.Option]time.Duration{
		updatecheck.Hourly: time.Hour,
		updatecheck.Daily:  24 * time.Hour,
	} {
		h := newPeriodicCheckHarness(option)
		h.check.reschedule()
		c.Equal(1, len(h.ticks), "%s must schedule exactly one tick", option.Key())
		c.Equal(want, h.ticks[0].after, "%s must schedule its tick at its own interval", option.Key())
		c.True(h.check.pending, "%s must leave a tick pending", option.Key())
	}
}

// A tick that ran the work twice would double the checks with every repetition, and one that failed to reschedule
// would silently turn the setting into a single check.
func TestPeriodicCheckTickRunsTheWorkAndSchedulesTheNext(t *testing.T) {
	c := check.New(t)
	h := newPeriodicCheckHarness(updatecheck.Hourly)
	h.check.reschedule()
	c.Equal(1, len(h.ticks))
	h.ticks[0].fire()
	c.Equal(1, h.runs, "the work must run exactly once per tick")
	c.Equal(2, len(h.ticks), "the tick must schedule exactly one successor")
	c.Equal(time.Hour, h.ticks[1].after)
	h.ticks[1].fire()
	c.Equal(2, h.runs)
	c.Equal(3, len(h.ticks))
}

// Scheduled ticks can't be canceled, so without the generation check a switch from hourly to daily would leave the
// hourly tick to fire anyway, and the two chains would then run side by side.
func TestPeriodicCheckRetiresTicksFromAnOldSetting(t *testing.T) {
	c := check.New(t)
	h := newPeriodicCheckHarness(updatecheck.Hourly)
	h.check.reschedule()
	stale := h.ticks[0]
	h.option = updatecheck.Daily
	h.check.reschedule()
	c.Equal(2, len(h.ticks))
	c.Equal(24*time.Hour, h.ticks[1].after)
	stale.fire()
	c.Equal(0, h.runs, "a tick from the previous setting must not run the work")
	c.Equal(2, len(h.ticks), "a tick from the previous setting must not schedule anything")
}

// A tick was already on its way when the user turned the checks off, and must not keep the chain going.
func TestPeriodicCheckStopsWhenSwitchedToNever(t *testing.T) {
	c := check.New(t)
	h := newPeriodicCheckHarness(updatecheck.Hourly)
	h.check.reschedule()
	stale := h.ticks[0]
	h.option = updatecheck.Never
	h.check.reschedule()
	c.False(h.check.pending)
	stale.fire()
	c.Equal(0, h.runs, "a tick that arrives after the checks were turned off must not run the work")
	c.Equal(1, len(h.ticks), "a tick that arrives after the checks were turned off must not schedule anything")
}

// The settings dialog applies the settings every time it syncs or resets, so a reschedule that always began the wait
// again would push the next check out indefinitely for anyone who left that dialog open.
func TestPeriodicCheckKeepsTheClockRunningWhenNothingChanged(t *testing.T) {
	c := check.New(t)
	h := newPeriodicCheckHarness(updatecheck.Hourly)
	h.check.reschedule()
	pending := h.ticks[0]
	h.check.reschedule()
	h.check.reschedule()
	c.Equal(1, len(h.ticks), "re-applying an unchanged setting must not schedule another tick")
	pending.fire()
	c.Equal(1, h.runs, "the tick that was already pending must still be the live one")
	c.Equal(2, len(h.ticks))
}

// Every path that changes the setting reschedules, which retires the tick, so this is a backstop; but a backstop that
// ran one last check the user had just turned off would be worse than none.
func TestPeriodicCheckTickSkipsTheWorkWhenTheSettingNoLongerRepeats(t *testing.T) {
	c := check.New(t)
	for _, option := range []updatecheck.Option{updatecheck.Never, updatecheck.AtLaunch} {
		h := newPeriodicCheckHarness(updatecheck.Hourly)
		c.True(h.check.reschedule(), "an hourly setting must report that the checks repeat")
		c.Equal(1, len(h.ticks))
		h.option = option
		h.ticks[0].fire()
		c.Equal(0, h.runs, "a tick that finds the setting at %s must not run the work", option.Key())
		c.Equal(1, len(h.ticks), "a tick that finds the setting at %s must not schedule another", option.Key())
		c.False(h.check.pending)
		c.False(h.check.reschedule(), "a setting of %s must report that the checks don't repeat", option.Key())
	}
}

// Switching on from Never runs the work immediately; nothing else does -- not starting the schedule at launch, whose
// check is made separately, and not moving between settings that already check. Without this, turning the checks on
// left the Help menu insisting that they were off, and no check happened until the next launch or, for the repeating
// settings, the first tick an hour or a day later.
func TestPeriodicCheckRunsAtOnceWhenTurnedOn(t *testing.T) {
	c := check.New(t)
	h := newPeriodicCheckHarness(updatecheck.Never)
	h.check.reschedule()
	c.Equal(0, h.runs, "starting the schedule with the checks off must not run the work")

	h.option = updatecheck.AtLaunch
	h.check.reschedule()
	c.Equal(1, h.runs, "turning the checks on must run the work at once")
	c.Equal(0, len(h.ticks), "checking at launch must still not schedule a repeating check")

	h.option = updatecheck.Hourly
	h.check.reschedule()
	c.Equal(1, h.runs, "moving between settings that check must not run the work again")
	c.Equal(1, len(h.ticks))

	h.option = updatecheck.Never
	h.check.reschedule()
	c.Equal(1, h.runs, "turning the checks off must not run the work")

	h.option = updatecheck.Daily
	h.check.reschedule()
	c.Equal(2, h.runs, "turning the checks on again must run the work again")
	c.Equal(2, len(h.ticks))
	c.Equal(24*time.Hour, h.ticks[1].after, "...as well as scheduling the repeating check")

	h.check.reschedule()
	c.Equal(2, h.runs, "re-applying an unchanged setting must not run the work")

	for _, option := range []updatecheck.Option{updatecheck.AtLaunch, updatecheck.Hourly, updatecheck.Daily} {
		fresh := newPeriodicCheckHarness(option)
		fresh.check.reschedule()
		c.Equal(0, fresh.runs, "starting the schedule with %s must not run the work", option.Key())
	}
}

// A development build never looks for updates, so it must not schedule repeating checks that would only wake up to say
// so again; a release build follows the setting as written.
func TestAppUpdateCheckFrequencyForDevelopmentBuilds(t *testing.T) {
	c := check.New(t)
	gs, _ := prepareUpdateCheckSettings(t)

	swapForTest(t, &xos.AppVersion, "0.0")
	for option, want := range map[updatecheck.Option]updatecheck.Option{
		updatecheck.Never:    updatecheck.Never,
		updatecheck.AtLaunch: updatecheck.AtLaunch,
		updatecheck.Hourly:   updatecheck.AtLaunch,
		updatecheck.Daily:    updatecheck.AtLaunch,
	} {
		gs.AppUpdateCheck = option
		c.Equal(want, appUpdateCheckFrequency(), "development build with %s", option.Key())
	}

	xos.AppVersion = "5.99.0"
	for _, option := range updatecheck.Options {
		gs.AppUpdateCheck = option
		c.Equal(option, appUpdateCheckFrequency(), "release build with %s", option.Key())
	}
}
