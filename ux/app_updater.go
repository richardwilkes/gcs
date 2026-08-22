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

type appUpdater struct {
	lock     sync.RWMutex
	result   string
	releases []gurps.Release
	updating bool
}

var appUpdate appUpdater

func (u *appUpdater) Reset() bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.updating {
		return false
	}
	u.result = fmt.Sprintf(i18n.Text("Checking for %s updates…"), xos.AppName)
	u.releases = nil
	u.updating = true
	return true
}

func (u *appUpdater) Result() (title string, releases []gurps.Release, updating bool) {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.result, u.releases, u.updating
}

func (u *appUpdater) SetResult(str string) {
	u.lock.Lock()
	u.result = str
	u.updating = false
	u.lock.Unlock()
}

func (u *appUpdater) SetReleases(releases []gurps.Release) {
	u.lock.Lock()
	u.result = fmt.Sprintf(i18n.Text("%s %s is available!"), xos.AppName, filterVersion(releases[0].Version))
	u.releases = releases
	u.updating = false
	u.lock.Unlock()
}

// CheckForAppUpdates initiates a fresh check for application updates.
func CheckForAppUpdates() {
	if xos.AppVersion == "0.0" {
		appUpdate.SetResult(fmt.Sprintf(i18n.Text("Development versions don't look for %s updates"), xos.AppName))
		return
	}
	if appUpdate.Reset() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			defer cancel()
			releases, err := gurps.LoadReleases(ctx, &http.Client{}, "richardwilkes", "", "gcs", xos.AppVersion,
				func(version, _ string) bool {
					// Don't bother showing changes from before 5.0.0, since those were the Java version
					return xstrings.NaturalLess(version, "5.0.0", true)
				}, false)
			if err != nil {
				appUpdate.SetResult(fmt.Sprintf(i18n.Text("Unable to access the %s update site"), xos.AppName))
				errs.Log(err)
				return
			}
			if len(releases) == 0 || releases[0].Version == xos.AppVersion {
				appUpdate.SetResult(fmt.Sprintf(i18n.Text("No %s updates are available"), xos.AppName))
				return
			}
			appUpdate.SetReleases(releases)
			unison.InvokeTask(NotifyOfAppUpdate)
		}()
	}
}

// downloadPageResponse is the response code for the button that opens the download page rather than installing. The
// pre-defined codes are taken by Cancel and by the default action, so a third button needs one of its own.
const downloadPageResponse = unison.ModalResponseUserBase

// NotifyOfAppUpdate notifies the user of the available update.
func NotifyOfAppUpdate() {
	title, releases, _ := appUpdate.Result()
	if releases == nil {
		return
	}
	// Work out whether this installation can update itself before the dialog is built, so that the choice offered
	// matches what is actually possible and the user is never told an update is being installed only to be refused
	// after thirty megabytes have been downloaded.
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
		if err = xos.OpenBrowser("https://" + WebSiteDomain); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to open web page for download"), err)
		}
	}
}

// planAppUpdate checks whether this installation can replace itself with the given release. It returns either a plan to
// do so, or a message explaining why it cannot, which the dialog shows alongside the release notes.
func planAppUpdate(release *gurps.Release) (plan *updater.Plan, unavailableMsg string) {
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

// blockerMessage explains, in the user's language, why an update cannot be installed automatically, and where possible
// what they can do about it. The updater deals in stable identifiers rather than messages precisely so that this
// translation happens here, at the point of display.
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
		// Installed successfully; only what came after it did not. Nothing here needs the user's attention badly
		// enough to interrupt them, since they are looking at the new version already.
		return
	}
	unison.WarningDialogWithMessage(fmt.Sprintf(i18n.Text("%s was not updated"), xos.AppName),
		xstrings.Wrap("", outcomeMessage(outcome.Reason), 100))
}

// outcomeMessage explains why an update that had already been prepared was not applied. The result is wrapped by the
// caller rather than here, since a dialog is the only thing that shows it and the wrapping belongs with the width the
// dialog wants.
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
func AppUpdateResult() (title string, releases []gurps.Release, updating bool) {
	return appUpdate.Result()
}
