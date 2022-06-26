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

package menus

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/updates"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var (
	// SponsorGCSDevelopment opens the web site for sponsoring GCS development.
	SponsorGCSDevelopment *unison.Action
	// MakeDonation opens the web site for make a donation.
	MakeDonation *unison.Action
	// UpdateAppStatus shows the status of the last app update check.
	UpdateAppStatus *unison.Action
	// CheckForAppUpdates requests another check for app updates.
	CheckForAppUpdates *unison.Action
	// ReleaseNotes opens the release notes.
	ReleaseNotes *unison.Action
	// License opens the license.
	License *unison.Action
	// WebSite opens the GCS web site.
	WebSite *unison.Action
	// MailingList opens the GCS mailing list site.
	MailingList *unison.Action
)

func registerHelpMenuActions() {
	SponsorGCSDevelopment = &unison.Action{
		ID:    constants.SponsorGCSDevelopmentItemID,
		Title: fmt.Sprintf(i18n.Text("Sponsor %s Development"), cmdline.AppName),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://github.com/sponsors/richardwilkes")
		},
	}
	MakeDonation = &unison.Action{
		ID:    constants.MakeDonationItemID,
		Title: fmt.Sprintf(i18n.Text("Make a One-time Donation for %s Development"), cmdline.AppName),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://paypal.me/GURPSCharacterSheet")
		},
	}
	UpdateAppStatus = &unison.Action{
		ID: constants.UpdateAppStatusItemID,
		EnabledCallback: func(action *unison.Action, mi any) bool {
			title, releases, updating := updates.AppUpdateResult()
			action.Title = title
			if menuItem, ok := mi.(unison.MenuItem); ok {
				menuItem.SetTitle(title)
			}
			return !updating && releases != nil
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if _, releases, updating := updates.AppUpdateResult(); !updating && releases != nil {
				updates.NotifyOfAppUpdate()
			}
		},
	}
	CheckForAppUpdates = &unison.Action{
		ID:    constants.CheckForAppUpdatesItemID,
		Title: fmt.Sprintf(i18n.Text("Check for %s updates"), cmdline.AppName),
		EnabledCallback: func(action *unison.Action, mi any) bool {
			_, releases, updating := updates.AppUpdateResult()
			return !updating && releases == nil
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			settings.Global().LastSeenGCSVersion = ""
			updates.CheckForAppUpdates()
		},
	}
	ReleaseNotes = &unison.Action{
		ID:    constants.ReleaseNotesItemID,
		Title: i18n.Text("Release Notes"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://github.com/richardwilkes/gcs/releases")
		},
	}
	License = &unison.Action{
		ID:    constants.ReleaseNotesItemID,
		Title: i18n.Text("License"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://github.com/richardwilkes/gcs/blob/master/LICENSE")
		},
	}
	WebSite = &unison.Action{
		ID:    constants.WebSiteItemID,
		Title: i18n.Text("Web Site"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://" + constants.WebSiteDomain)
		},
	}
	MailingList = &unison.Action{
		ID:    constants.MailingListItemID,
		Title: i18n.Text("Mailing Lists"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://groups.io/g/gcs")
		},
	}
}

func setupHelpMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.HelpMenuID)
	m.InsertItem(-1, SponsorGCSDevelopment.NewMenuItem(f))
	m.InsertItem(-1, MakeDonation.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, UpdateAppStatus.NewMenuItem(f))
	m.InsertItem(-1, CheckForAppUpdates.NewMenuItem(f))
	m.InsertItem(-1, ReleaseNotes.NewMenuItem(f))
	m.InsertItem(-1, License.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, WebSite.NewMenuItem(f))
	m.InsertItem(-1, MailingList.NewMenuItem(f))
}

func showWebPage(uri string) {
	if err := desktop.Open(uri); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to open link"), err)
	}
}
