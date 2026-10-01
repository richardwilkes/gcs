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
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/gcs/v5/updater"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/behavior"
)

// appUpdater holds what is known about available application updates. A visible check (asked for by the user or run at
// launch) blanks what is known and sets updating, so the menu and toolbar button show a check under way. A quiet check
// (run by the repeating schedule) leaves the previous answer on display until it has a better one, so a failed or
// unchanged background check never takes away an update the user was already told about.
type appUpdater struct {
	lock      sync.RWMutex
	frequency func() updatecheck.Option // nil means the general settings' AppUpdateCheck; tests inject a value
	result    string
	releases  []library.Release
	updating  bool
	quiet     bool // a quiet check is in flight
	seq       int  // bumped by every visible check, so a quiet result that arrives late can be recognized as stale
}

var appUpdate appUpdater

// Reset marks the start of a visible check, returning false if one is already running. Bumping seq makes any quiet
// check still in flight discard its result.
func (u *appUpdater) Reset() bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.updating {
		return false
	}
	u.result = checkingForAppUpdatesText()
	u.releases = nil
	u.updating = true
	u.seq++
	return true
}

// Result returns what is currently known. Until a check records something, the title says why: a quiet check is under
// way, the checks are off, or none has run yet. The Help menu shows the title verbatim, so it must never be blank.
func (u *appUpdater) Result() (title string, releases []library.Release, updating bool) {
	u.lock.RLock()
	defer u.lock.RUnlock()
	if u.result == "" {
		return u.uncheckedTitleLocked(), u.releases, u.updating
	}
	return u.result, u.releases, u.updating
}

// Checking returns true while a check of either kind is in flight, unlike Result, which reports only a visible check as
// updating. The Help menu's check item uses it to avoid starting a second request.
func (u *appUpdater) Checking() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.updating || u.quiet
}

// uncheckedTitleLocked returns the title for the state in which no check has recorded anything. The lock must already
// be held.
func (u *appUpdater) uncheckedTitleLocked() string {
	switch {
	case u.quiet:
		return checkingForAppUpdatesText()
	case u.option() == updatecheck.Never:
		return fmt.Sprintf(i18n.Text("Automatic %s update checks are off"), xos.AppName)
	default:
		return fmt.Sprintf(i18n.Text("No %s update check has run yet"), xos.AppName)
	}
}

func (u *appUpdater) option() updatecheck.Option {
	if u.frequency != nil {
		return u.frequency()
	}
	return gurps.GlobalSettings().General.AppUpdateCheck
}

// SetResult records a check outcome that has no releases to offer and marks any visible check as finished.
func (u *appUpdater) SetResult(str string) {
	u.lock.Lock()
	u.result = str
	u.updating = false
	u.lock.Unlock()
}

// SetReleases records the releases a visible check found.
func (u *appUpdater) SetReleases(releases []library.Release) {
	u.lock.Lock()
	u.setReleasesLocked(releases)
	u.lock.Unlock()
}

// setReleasesLocked records the releases an update check found. The lock must already be held.
func (u *appUpdater) setReleasesLocked(releases []library.Release) {
	u.result = fmt.Sprintf(i18n.Text("%s %s is available!"), xos.AppName, filterVersion(releases[0].Version))
	u.releases = releases
	u.updating = false
}

func noAppUpdatesText() string {
	return fmt.Sprintf(i18n.Text("No %s updates are available"), xos.AppName)
}

func checkingForAppUpdatesText() string {
	return fmt.Sprintf(i18n.Text("Checking for %s updates…"), xos.AppName)
}

func unableToAccessAppUpdateSiteText() string {
	return fmt.Sprintf(i18n.Text("Unable to access the %s update site"), xos.AppName)
}

func devVersionAppUpdateText() string {
	return fmt.Sprintf(i18n.Text("Development versions don't look for %s updates"), xos.AppName)
}

// beginQuiet claims the right to run a quiet check, returning the sequence number to hand back to finishQuiet. It
// refuses while a visible check is running, since that check's answer is the one the user is waiting on, and while
// another quiet check is already in flight.
func (u *appUpdater) beginQuiet() (seq int, ok bool) {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.updating || u.quiet {
		return 0, false
	}
	u.quiet = true
	return u.seq, true
}

// finishQuiet records the outcome of a quiet check. A result whose sequence number no longer matches is discarded,
// since a visible check has started since. A failed check (the caller logs the error) leaves what is known untouched,
// so a network hiccup can't erase an update the user was already told about; when nothing was known, the failure is
// recorded so the Help menu says the site couldn't be reached rather than that no check has run.
func (u *appUpdater) finishQuiet(seq int, releases []library.Release, err error) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.quiet = false
	if seq != u.seq {
		return
	}
	if err != nil {
		if u.result == "" {
			u.result = unableToAccessAppUpdateSiteText()
		}
		return
	}
	if len(releases) == 0 || releases[0].Version == xos.AppVersion {
		u.result = noAppUpdatesText()
		u.releases = nil
		return
	}
	u.setReleasesLocked(releases)
}

// loadAppReleases retrieves the releases newer than the running version. With nothing newer, it may hold just the
// running version's own release.
func loadAppReleases(ctx context.Context) ([]library.Release, error) {
	return library.LoadReleases(ctx, &http.Client{}, "richardwilkes", "", "gcs", xos.AppVersion,
		func(version, _ string) bool {
			// Don't bother showing changes from before 5.0.0, since those were the Java version
			return xstrings.NaturalLess(version, "5.0.0", true)
		}, false)
}

// CheckForAppUpdates starts a visible check for application updates: the state shows a check under way, and a release
// not yet shown opens the notification dialog. A development build has no release to compare against, so it reports
// that instead of checking.
func CheckForAppUpdates() {
	if updater.IsDevVersion(xos.AppVersion) {
		appUpdate.SetResult(devVersionAppUpdateText())
		unison.InvokeTask(SyncAppUpdateButton)
		return
	}
	if appUpdate.Reset() {
		unison.InvokeTask(SyncAppUpdateButton)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			defer cancel()
			releases, err := loadAppReleases(ctx)
			if err != nil {
				appUpdate.SetResult(unableToAccessAppUpdateSiteText())
				errs.Log(err)
				unison.InvokeTask(SyncAppUpdateButton)
				return
			}
			if len(releases) == 0 || releases[0].Version == xos.AppVersion {
				appUpdate.SetResult(noAppUpdatesText())
				unison.InvokeTask(SyncAppUpdateButton)
				return
			}
			appUpdate.SetReleases(releases)
			unison.InvokeTask(func() {
				SyncAppUpdateButton()
				if shouldShowAppUpdateDialog(releases[0].Version, gurps.GlobalSettings().LastSeenGCSVersion) {
					NotifyOfAppUpdate()
				}
			})
		}()
	}
}

// shouldShowAppUpdateDialog reports whether a release warrants interrupting the user with the notification dialog.
// LastSeenGCSVersion is written by NotifyOfAppUpdate just before it goes modal and is cleared by the Help menu's check
// item, so the dialog opens the first time a release is seen and whenever the user asks for a check, while later
// launches that turn up an already declined release say so with the toolbar button alone.
func shouldShowAppUpdateDialog(version, lastSeen string) bool {
	return version != lastSeen
}

// checkForAppUpdatesQuietly checks for application updates for the repeating schedule. It never opens the dialog, and a
// failed check leaves what is known in place.
func checkForAppUpdatesQuietly() {
	if updater.IsDevVersion(xos.AppVersion) {
		// Development versions never look for updates; saying so, as the visible check does, beats a status claiming
		// that a check is still to come.
		appUpdate.SetResult(devVersionAppUpdateText())
		return
	}
	seq, ok := appUpdate.beginQuiet()
	if !ok {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()
		releases, err := loadAppReleases(ctx)
		if err != nil {
			errs.Log(err)
		}
		appUpdate.finishQuiet(seq, releases, err)
		unison.InvokeTask(SyncAppUpdateButton)
	}()
}

// downloadPageResponse is the Download Page button's response code; Cancel and OK already use the predefined ones.
const downloadPageResponse = unison.ModalResponseUserBase

// NotifyOfAppUpdate notifies the user of the available update.
func NotifyOfAppUpdate() {
	title, releases, _ := appUpdate.Result()
	if releases == nil {
		return
	}
	// Check whether this installation can update itself before building the dialog, so it never offers an install
	// that would be refused after a large download.
	plan, unavailableMsg := planAppUpdate(&releases[0])

	var buffer strings.Builder
	fmt.Fprintf(&buffer, "# %s\n", title)
	for i, rel := range releases {
		if i != 0 {
			buffer.WriteString("---\n")
		}
		fmt.Fprintf(&buffer, "## Release Notes for %s %s\n", xos.AppName, filterVersion(rel.Version))
		buffer.WriteString(rel.Notes)
		buffer.WriteByte('\n')
	}
	if unavailableMsg != "" {
		fmt.Fprintf(&buffer, "---\n%s\n", unavailableMsg)
	}

	md := unison.NewMarkdown(true)
	md.SetBorder(unison.NewEmptyBorder(unison.StdInsets()))
	md.SetContent(buffer.String(), 0)

	scroll := unison.NewScrollPanel()
	scroll.SetContent(md, behavior.Unmodified, behavior.Unmodified)

	buttons := []*unison.DialogButtonInfo{
		unison.NewCancelButtonInfo(),
		{Title: i18n.Text("Download Page"), ResponseCode: downloadPageResponse},
	}
	if plan != nil {
		buttons = append(buttons, unison.NewOKButtonInfoWithTitle(i18n.Text("Install & Restart")))
	}

	dialog, err := unison.NewDialog(
		&unison.DrawableSVG{
			SVG:  svg.Download,
			Size: geom.NewSize(48, 48),
		},
		unison.DefaultLabelTheme.OnBackgroundInk, scroll, buttons,
	)
	if err != nil {
		errs.Log(err)
		return
	}
	gurps.GlobalSettings().LastSeenGCSVersion = releases[0].Version
	switch dialog.RunModal() {
	case unison.ModalResponseOK:
		InitiateAppUpdate(plan)
	case downloadPageResponse:
		if err = unison.OpenBrowser("https://" + WebSiteDomain); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to open web page for download"), err)
		}
	}
}

// planAppUpdate checks whether this installation can replace itself with the given release. It returns either a plan to
// do so, or a message explaining why it cannot, which the dialog shows alongside the release notes.
func planAppUpdate(release *library.Release) (plan *updater.Plan, unavailableMsg string) {
	assets := make([]updater.Asset, len(release.Assets))
	for i := range release.Assets {
		assets[i] = updater.Asset{
			Name:   release.Assets[i].Name,
			URL:    release.Assets[i].URL,
			SHA256: release.Assets[i].SHA256(),
			Size:   release.Assets[i].Size,
		}
	}
	plan, err := updater.CurrentPreflight(release.Version, assets)
	if err == nil {
		return plan, ""
	}
	errs.Log(err)
	if unavailable, ok := errors.AsType[*updater.Unavailable](err); ok {
		return nil, blockerMessage(unavailable.Blocker)
	}
	return nil, fmt.Sprintf(i18n.Text("%s can't install this update automatically."), xos.AppName)
}

// blockerMessage explains, in the user's language, why an update cannot be installed automatically and, where possible,
// what to do about it. The updater reports stable identifiers so that translation happens here, at display time.
func blockerMessage(blocker updater.Blocker) string {
	switch blocker {
	case updater.BlockerDevBuild:
		return i18n.Text("Development builds can't be updated automatically.")
	case updater.BlockerRenamedExecutable:
		return fmt.Sprintf(i18n.Text("This copy of %s has been renamed, so it can't be updated automatically."),
			xos.AppName)
	case updater.BlockerNotABundle:
		return fmt.Sprintf(i18n.Text("This copy of %s isn't installed as an application, so it can't be updated automatically."),
			xos.AppName)
	case updater.BlockerTranslocated:
		return fmt.Sprintf(i18n.Text("%s is running from a temporary copy. Move it to your Applications folder and open it from there to enable automatic updates."),
			xos.AppName)
	case updater.BlockerPackageManaged:
		return fmt.Sprintf(i18n.Text("%s was installed by a package manager, which should be used to update it."),
			xos.AppName)
	case updater.BlockerHomebrew:
		return i18n.Text("This copy was installed by Homebrew. Run `brew upgrade --cask gcs` to update it.")
	case updater.BlockerReadOnly:
		return fmt.Sprintf(i18n.Text("%s can't write to the folder it's installed in, so it can't update itself. Installing it somewhere you have permission to write, such as your Applications folder, enables automatic updates."),
			xos.AppName)
	case updater.BlockerNoAsset:
		return i18n.Text("This release doesn't include a build for this system.")
	case updater.BlockerNoDigest:
		return i18n.Text("This release can't be verified automatically, so it has to be installed by hand.")
	default:
		return fmt.Sprintf(i18n.Text("%s can't install this update automatically."), xos.AppName)
	}
}

// ReportAppUpdateOutcome settles up after an update that was applied while the application was not running, clearing
// away what it left behind and telling the user if anything went wrong. Success is silent: the version shown in the
// About box is confirmation enough.
func ReportAppUpdateOutcome() {
	outcome := updater.ReportAndSweep()
	if outcome == nil || outcome.Reason == updater.ReasonNone || outcome.Reason == updater.ReasonAbandoned {
		return
	}
	if outcome.Applied {
		// Installed; only a later step, most likely the relaunch, failed. That isn't worth interrupting a user who is
		// already running the new version.
		return
	}
	unison.WarningDialogWithMessage(fmt.Sprintf(i18n.Text("%s was not updated"), xos.AppName),
		xstrings.Wrap("", outcomeMessage(outcome.Reason), 100))
}

// outcomeMessage explains why a prepared update was not applied. The caller wraps the result to its dialog's width.
func outcomeMessage(reason updater.Reason) string {
	switch reason {
	case updater.ReasonPredecessorRunning:
		return fmt.Sprintf(i18n.Text("The previous copy of %s did not finish quitting, so nothing was changed. Try the update again."),
			xos.AppName)
	case updater.ReasonSwapFailed:
		return fmt.Sprintf(i18n.Text("The update could not be installed, so the previous version was kept. Downloading it from the %s web site and installing it by hand will work."),
			xos.AppName)
	case updater.ReasonVersionMismatch:
		return i18n.Text("The update reported success but a different version is running. The previous version has been kept.")
	default:
		return fmt.Sprintf(i18n.Text("The update could not be installed. Downloading it from the %s web site and installing it by hand will work."),
			xos.AppName)
	}
}

// AppUpdateResult returns the current results of any outstanding app update check.
func AppUpdateResult() (title string, releases []library.Release, updating bool) {
	return appUpdate.Result()
}

// AppUpdateCheckInProgress returns true while an app update check of either kind, visible or quiet, is in flight.
func AppUpdateCheckInProgress() bool {
	return appUpdate.Checking()
}
