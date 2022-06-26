/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package updates

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

type appUpdater struct {
	lock     sync.RWMutex
	result   string
	releases []library.Release
	updating bool
}

var appUpdate appUpdater

func (u *appUpdater) Reset() bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.updating {
		return false
	}
	u.result = fmt.Sprintf(i18n.Text("Checking for %s updates…"), cmdline.AppName)
	u.releases = nil
	u.updating = true
	return true
}

func (u *appUpdater) Result() (title string, releases []library.Release, updating bool) {
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

func (u *appUpdater) SetReleases(releases []library.Release) {
	u.lock.Lock()
	u.result = fmt.Sprintf(i18n.Text("%s v%s is available!"), cmdline.AppName, releases[0].Version)
	u.releases = releases
	u.updating = false
	u.lock.Unlock()
}

// CheckForAppUpdates initiates a fresh check for application updates.
func CheckForAppUpdates() {
	if cmdline.AppVersion == "0.0" {
		appUpdate.SetResult(fmt.Sprintf(i18n.Text("Development versions don't look for %s updates"), cmdline.AppName))
		return
	}
	if appUpdate.Reset() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			defer cancel()
			releases, err := library.LoadReleases(ctx, &http.Client{}, "richardwilkes", "gcs", cmdline.AppVersion,
				func(version, notes string) bool {
					// Don't bother showing changes from before 5.0.0, since those were the Java version
					return txt.NaturalLess(version, "5.0.0", true)
				})
			if err != nil {
				appUpdate.SetResult(fmt.Sprintf(i18n.Text("Unable to access the %s update site"), cmdline.AppName))
				jot.Error(err)
				return
			}
			if len(releases) == 0 || releases[0].Version == cmdline.AppVersion {
				appUpdate.SetResult(fmt.Sprintf(i18n.Text("No %s updates are available"), cmdline.AppName))
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
			fmt.Fprintf(&buffer, "## Release Notes for %s v%s\n", cmdline.AppName, rel.Version)
			buffer.WriteString(rel.Notes)
			buffer.WriteByte('\n')
		}

		scroll := unison.NewScrollPanel()
		scroll.SetContent(convertMarkdownToPanel(buffer.String(), 900), unison.UnmodifiedBehavior,
			unison.UnmodifiedBehavior)

		dialog, err := unison.NewDialog(
			&unison.DrawableSVG{
				SVG:  res.DownloadSVG,
				Size: unison.NewSize(48, 48),
			},
			unison.DefaultLabelTheme.OnBackgroundInk, scroll,
			[]*unison.DialogButtonInfo{
				unison.NewCancelButtonInfo(),
				unison.NewOKButtonInfoWithTitle(i18n.Text("Download")),
			})
		if err != nil {
			jot.Error(err)
			return
		}
		settings.Global().LastSeenGCSVersion = releases[0].Version
		if dialog.RunModal() == unison.ModalResponseOK {
			if err = desktop.Open("https://" + constants.WebSiteDomain); err != nil {
				unison.ErrorDialogWithError(i18n.Text("Unable to open web page for download"), err)
			}
		}
	}
}

// AppUpdateResult returns the current results of any outstanding app update check.
func AppUpdateResult() (title string, releases []library.Release, updating bool) {
	return appUpdate.Result()
}

func convertMarkdownToPanel(text string, maxWidth float32) *unison.Panel {
	hdrFD := unison.DefaultLabelTheme.Font.Descriptor()
	hdrFD.Weight = unison.BoldFontWeight
	sizes := []float32{hdrFD.Size * 2, hdrFD.Size * 3 / 2, hdrFD.Size * 5 / 4, hdrFD.Size}
	otherDecoration := &unison.TextDecoration{Font: unison.DefaultLabelTheme.Font}
	bulletWidth := maxWidth - (otherDecoration.Font.SimpleWidth("•") + unison.StdHSpacing)
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: 0,
	})
	for _, line := range strings.Split(text, "\n") {
		found := false
		for i := range sizes {
			if strings.HasPrefix(line, strings.Repeat("#", i+1)+" ") {
				hdrFD.Size = sizes[i]
				font := hdrFD.Font()
				block := unison.NewTextWrappedLines(line[i+2:], &unison.TextDecoration{Font: font}, maxWidth)
				for j, chunk := range block {
					label := unison.NewLabel()
					label.Font = font
					label.Text = chunk.String()
					if j == len(block)-1 {
						label.SetBorder(unison.NewEmptyBorder(unison.Insets{Bottom: sizes[i] / 2}))
					}
					label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
					panel.AddChild(label)
				}
				found = true
				break
			}
		}
		if found {
			continue
		}
		switch {
		case line == "---":
			hr := unison.NewSeparator()
			hr.SetBorder(unison.NewEmptyBorder(unison.NewVerticalInsets(otherDecoration.Font.Size())))
			hr.SetLayoutData(&unison.FlexLayoutData{
				HSpan:  2,
				HAlign: unison.FillAlignment,
				HGrab:  true,
			})
			panel.AddChild(hr)
		case strings.HasPrefix(line, "- "):
			block := unison.NewTextWrappedLines(line[2:], otherDecoration, bulletWidth)
			for j, chunk := range block {
				label := unison.NewLabel()
				label.Font = otherDecoration.Font
				if j == 0 {
					label.Text = "•"
				}
				panel.AddChild(label)

				label = unison.NewLabel()
				label.Font = otherDecoration.Font
				label.Text = chunk.String()
				panel.AddChild(label)
			}
		default:
			block := unison.NewTextWrappedLines(line, otherDecoration, maxWidth)
			for _, chunk := range block {
				label := unison.NewLabel()
				label.Font = otherDecoration.Font
				label.Text = chunk.String()
				label.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
				panel.AddChild(label)
			}
		}
	}
	return panel
}
