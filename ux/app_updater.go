// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
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

// NotifyOfAppUpdate notifies the user of the available update.
func NotifyOfAppUpdate() {
	if title, releases, _ := appUpdate.Result(); releases != nil {
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

		md := unison.NewMarkdown(true)
		md.SetBorder(unison.NewEmptyBorder(unison.StdInsets()))
		md.SetContent(buffer.String(), 0)

		scroll := unison.NewScrollPanel()
		scroll.SetContent(md, behavior.Unmodified, behavior.Unmodified)

		dialog, err := unison.NewDialog(
			&unison.DrawableSVG{
				SVG:  svg.Download,
				Size: geom.NewSize(48, 48),
			},
			unison.DefaultLabelTheme.OnBackgroundInk, scroll,
			[]*unison.DialogButtonInfo{
				unison.NewCancelButtonInfo(),
				unison.NewOKButtonInfoWithTitle(i18n.Text("Download")),
			})
		if err != nil {
			errs.Log(err)
			return
		}
		gurps.GlobalSettings().LastSeenGCSVersion = releases[0].Version
		if dialog.RunModal() == unison.ModalResponseOK {
			if err = xos.OpenBrowser("https://" + WebSiteDomain); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to open web page for download"), err)
			}
		}
	}
}

// AppUpdateResult returns the current results of any outstanding app update check.
func AppUpdateResult() (title string, releases []gurps.Release, updating bool) {
	return appUpdate.Result()
}
