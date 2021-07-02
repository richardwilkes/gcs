/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.library;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.library.LibraryUpdater;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Release;
import com.trollworks.gcs.utility.UpdateChecker;
import com.trollworks.gcs.utility.Version;

import java.awt.event.ActionEvent;

public class LibraryUpdateCommand extends Command {
    private Library mLibrary;

    public LibraryUpdateCommand(Library library) {
        super(library.getTitle(), "lib:" + library.getKey());
        mLibrary = library;
    }

    @Override
    public void adjust() {
        Release upgrade = mLibrary.getAvailableUpgrade();
        String  title   = mLibrary.getTitle();
        if (upgrade == null) {
            setTitle(String.format(I18n.text("Checking for updates to %s"), title));
            setEnabled(false);
            return;
        }
        if (upgrade.unableToAccessRepo()) {
            setTitle(String.format(I18n.text("Unable to access the %s repo"), title));
            setEnabled(false);
            return;
        }
        if (!upgrade.hasUpdate()) {
            setTitle(String.format(I18n.text("No releases available for %s"), title));
            setEnabled(false);
            return;
        }
        Version versionOnDisk    = mLibrary.getVersionOnDisk();
        Version availableVersion = upgrade.getVersion();
        if (availableVersion.equals(versionOnDisk)) {
            setTitle(String.format(I18n.text("%s is up to date (re-download v%s)"), title, versionOnDisk));
        } else {
            setTitle(String.format(I18n.text("Update %s to v%s"), title, availableVersion));
        }
        setEnabled(!UIUtilities.inModalState());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Release availableUpgrade = mLibrary.getAvailableUpgrade();
        if (availableUpgrade != null) {
            askUserToUpdate(mLibrary, availableUpgrade);
        }
    }

    public static void askUserToUpdate(Library library, Release release) {
        if (UpdateChecker.presentUpdateToUser(String.format(I18n.text("%s v%s is available!"),
                library.getTitle(), release.getVersion()), I18n.text("""
                NOTE: Existing content for this library will be removed and replaced. Content in other libraries will not be modified.

                """) + release.getNotes()).getResult() == Modal.OK) {
            LibraryUpdater.download(library, release);
        }
    }
}
