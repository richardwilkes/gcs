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
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/updater"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
)

// periodicCheck repeats work on the UI thread at the interval an update-check setting asks for, and runs it at once
// when the checks are turned on. A scheduled tick can't be canceled, so each one carries the generation it was
// scheduled in; rescheduling bumps the generation, which retires anything already pending. reschedule and tick must
// only be called on the UI thread, which is also where unison.InvokeTaskAfter delivers the ticks.
type periodicCheck struct {
	frequency func() updatecheck.Option
	run       func()
	schedule  func(f func(), after time.Duration) // nil means unison.InvokeTaskAfter; tests inject a recorder
	gen       int
	pending   bool               // a tick is scheduled and has not fired
	scheduled updatecheck.Option // the option the pending tick was scheduled for
	applied   updatecheck.Option // the option in force the last time reschedule ran
	started   bool               // reschedule has run at least once
}

// reschedule brings the pending tick into line with the current setting, and runs the work at once if the checks have
// just been turned on. It returns true if the setting calls for repeating checks, which is to say if a tick is now
// pending. A pending tick that was already scheduled for the setting now in force is left alone, so that
// re-applying unchanged settings -- which the settings dialog does every time it syncs or resets -- doesn't push the
// next check further out each time. Switching from Never to anything else is the one change that runs the work
// immediately: Never is the only setting under which no check has been made this session, and a user who turns the
// checks on expects one to happen rather than having to wait an hour, a day, or until the next launch for the first.
// The first call, which is how the schedule is started at launch, doesn't count as turning the checks on, since the
// launch check is made separately.
func (p *periodicCheck) reschedule() bool {
	opt := p.frequency()
	turnedOn := p.started && p.applied == updatecheck.Never && opt != updatecheck.Never
	p.started = true
	p.applied = opt
	if !p.pending || opt != p.scheduled {
		p.gen++ // Retires any tick already pending, since it no longer matches the generation
		p.pending = false
		if interval := opt.Interval(); interval > 0 {
			gen := p.gen
			p.pending = true
			p.scheduled = opt
			schedule := p.schedule
			if schedule == nil {
				schedule = unison.InvokeTaskAfter
			}
			schedule(func() { p.tick(gen) }, interval)
		}
	}
	if turnedOn {
		p.run()
	}
	return p.pending
}

// tick runs the work and sets up the next repetition. Ticks left over from a previous setting are ignored. The next
// tick is scheduled before the work runs, both so that the cadence is measured start-to-start and so that a panic
// inside the work -- which unison recovers on the UI thread -- doesn't silently end the chain. A tick that finds the
// setting no longer calls for repeating checks -- which every path that changes the setting reports through
// ApplyUpdateCheckSettings, so this is a backstop -- schedules nothing further and doesn't run the work either, since
// the user has just asked for that not to happen.
func (p *periodicCheck) tick(gen int) {
	if gen != p.gen {
		return
	}
	p.pending = false
	if p.reschedule() {
		p.run()
	}
}

var (
	appUpdateCheck = periodicCheck{
		frequency: appUpdateCheckFrequency,
		run:       checkForAppUpdatesQuietly,
	}
	libraryUpdateCheck = periodicCheck{
		frequency: func() updatecheck.Option { return gurps.GlobalSettings().General.LibraryUpdateCheck },
		run:       func() { gurps.GlobalSettings().Libraries.PerformUpdateChecks() },
	}
)

// appUpdateCheckFrequency returns the app update check setting as it applies to this build. A development build never
// looks for updates, since there is no release behind it to compare against, so there is no point in waking up every
// hour or day to say so again: for it, the repeating options are treated as checking only at launch. Turning the checks
// on still runs the one check that records why nothing is being looked for.
func appUpdateCheckFrequency() updatecheck.Option {
	opt := gurps.GlobalSettings().General.AppUpdateCheck
	if opt.Interval() > 0 && updater.IsDevVersion(xos.AppVersion) {
		return updatecheck.AtLaunch
	}
	return opt
}

// ApplyUpdateCheckSettings brings the repeating update checks into line with the current settings. Call it on the UI
// thread after anything that may have changed those settings, as well as once at startup to start the schedules.
func ApplyUpdateCheckSettings() {
	appUpdateCheck.reschedule()
	libraryUpdateCheck.reschedule()
	SyncAppUpdateButton()
}
