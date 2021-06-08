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
import com.trollworks.gcs.ui.MarkdownDocument;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Release;
import com.trollworks.gcs.utility.Version;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import javax.swing.JOptionPane;
import javax.swing.JTextPane;

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
            setEnabled(true);
            return;
        }
        setTitle(String.format(I18n.text("Update %s to v%s"), title, availableVersion));
        setEnabled(true);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Release availableUpgrade = mLibrary.getAvailableUpgrade();
        if (availableUpgrade != null) {
            askUserToUpdate(mLibrary, availableUpgrade);
        }
    }

    public static void askUserToUpdate(Library library, Release release) {
        JTextPane markdown = new JTextPane(new MarkdownDocument(I18n.text("NOTE: Existing content for this library will be removed and replaced. Content in other libraries will not be modified.\n\n" + release.getNotes())));
        markdown.setBorder(new EmptyBorder(4));
        Dimension size     = markdown.getPreferredSize();
        int       maxWidth = Math.min(600, WindowUtils.getMaximumWindowBounds().width * 3 / 2);
        if (size.width > maxWidth) {
            markdown.setSize(new Dimension(maxWidth, Short.MAX_VALUE));
            size = markdown.getPreferredSize();
            size.width = maxWidth;
            markdown.setPreferredSize(size);
        }
        markdown.setEditable(false);
        String update = I18n.text("Update");
        if (WindowUtils.showOptionDialog(null, new ScrollPanel(markdown), String.format(I18n.text("%s v%s is available!"), library.getTitle(), release.getVersion()), true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.QUESTION_MESSAGE, null, new String[]{update, I18n.text("Ignore")}, update) == JOptionPane.OK_OPTION) {
            LibraryUpdater.download(library, release);
        }
    }
}
