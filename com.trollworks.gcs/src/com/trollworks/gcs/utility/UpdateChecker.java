/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility;

import com.trollworks.gcs.GCS;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.library.LibraryUpdateCommand;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.MarkdownDocument;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.task.Tasks;

import java.awt.Desktop;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.net.URI;
import java.util.List;
import java.util.concurrent.TimeUnit;
import javax.swing.JOptionPane;
import javax.swing.JScrollPane;
import javax.swing.JTextPane;

/** Provides a background check for updates. */
public class UpdateChecker implements Runnable {
    private static String  APP_RESULT;
    private static String  APP_RELEASE_NOTES;
    private static boolean NEW_APP_VERSION_AVAILABLE;
    private        Mode    mMode;

    private enum Mode {CHECK, NOTIFY, DONE}

    /**
     * Initiates a check for updates.
     */
    public static void check() {
        Thread thread = new Thread(new UpdateChecker(), UpdateChecker.class.getSimpleName());
        thread.setPriority(Thread.NORM_PRIORITY);
        thread.setDaemon(true);
        thread.start();
    }

    /** @return Whether a new app version is available. */
    public static synchronized boolean isNewAppVersionAvailable() {
        return NEW_APP_VERSION_AVAILABLE;
    }

    /** @return The result of the new app check. */
    public static synchronized String getAppResult() {
        return APP_RESULT != null ? APP_RESULT : I18n.Text("Checking for GCS updates…");
    }

    public static synchronized String getAppReleaseNotes() {
        return APP_RELEASE_NOTES;
    }

    private static synchronized void setAppResult(String result, String releaseNotes, boolean available) {
        APP_RESULT = result;
        APP_RELEASE_NOTES = releaseNotes;
        NEW_APP_VERSION_AVAILABLE = available;
    }

    /** Go to the update location on the web, if a new version is available. */
    public static void goToUpdate() {
        if (isNewAppVersionAvailable() && Desktop.isDesktopSupported()) {
            try {
                Desktop.getDesktop().browse(new URI(GCS.WEB_SITE));
            } catch (Exception exception) {
                WindowUtils.showError(null, exception.getMessage());
            }
        }
    }

    private UpdateChecker() {
        mMode = Mode.CHECK;
    }

    @Override
    public void run() {
        switch (mMode) {
        case CHECK:
            setAppResult(null, null, false);
            checkForAppUpdates();
            checkForLibraryUpdates();
            if (mMode == Mode.NOTIFY) {
                EventQueue.invokeLater(this);
            } else {
                mMode = Mode.DONE;
            }
            break;
        case NOTIFY:
            tryNotify();
            break;
        default:
            break;
        }
    }

    private void checkForAppUpdates() {
        if (GCS.VERSION.isZero()) {
            // Development version. Bail.
            setAppResult(I18n.Text("Development versions don't look for GCS updates"), null, false);
        } else {
            Version       minimum  = new Version(4, 17, 0);
            List<Release> releases = Release.load("richardwilkes", "gcs", GCS.VERSION, (version, notes) -> version.compareTo(minimum) >= 0);
            if (releases == null) {
                setAppResult(I18n.Text("Unable to access the GCS repo"), null, false);
                return;
            }
            int count = releases.size() - 1;
            if (count >= 0 && releases.get(count).getVersion().equals(GCS.VERSION)) {
                releases.remove(count);
            }
            if (releases.isEmpty()) {
                setAppResult(I18n.Text("GCS has no update available"), null, false);
            } else {
                Release     release   = new Release(releases);
                Preferences prefs     = Preferences.getInstance();
                Version     available = release.getVersion();
                setAppResult(String.format(I18n.Text("GCS v%s is available!"), available), release.getNotes(), true);
                if (available.compareTo(prefs.getLastSeenGCSVersion()) > 0) {
                    prefs.setLastSeenGCSVersion(available);
                    prefs.save();
                    mMode = Mode.NOTIFY;
                }
            }
        }
    }

    private void checkForLibraryUpdates() {
        for (Library lib : Library.LIBRARIES) {
            if (lib != Library.USER) {
                List<Release> releases = lib.checkForAvailableUpgrade();
                Version       lastSeen = lib.getLastSeen();
                lib.setAvailableUpgrade(releases);
                if (lib.getAvailableUpgrade().getVersion().compareTo(lastSeen) > 0) {
                    mMode = Mode.NOTIFY;
                }
            }
        }
    }

    private void tryNotify() {
        if (GCS.isNotificationAllowed()) {
            String update = I18n.Text("Update");
            mMode = Mode.DONE;
            if (isNewAppVersionAvailable()) {
                JTextPane   markdown = new JTextPane(new MarkdownDocument(getAppReleaseNotes()));
                Dimension   size     = markdown.getPreferredSize();
                JScrollPane scroller = new JScrollPane(markdown);
                int         maxWidth = Math.min(600, WindowUtils.getMaximumWindowBounds().width * 3 / 2);
                if (size.width > maxWidth) {
                    markdown.setSize(new Dimension(maxWidth, Short.MAX_VALUE));
                    size = markdown.getPreferredSize();
                    size.width = maxWidth;
                    markdown.setPreferredSize(size);
                }
                if (WindowUtils.showOptionDialog(null, scroller, getAppResult(), true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.QUESTION_MESSAGE, null, new String[]{update, I18n.Text("Ignore")}, update) == JOptionPane.OK_OPTION) {
                    goToUpdate();
                }
                return;
            }
            for (Library lib : Library.LIBRARIES) {
                if (lib != Library.USER) {
                    Release release = lib.getAvailableUpgrade();
                    if (release != null && !release.unableToAccessRepo() && release.hasUpdate() && !release.getVersion().equals(lib.getVersionOnDisk())) {
                        LibraryUpdateCommand.askUserToUpdate(lib, release);
                    }
                }
            }
        } else {
            Tasks.scheduleOnUIThread(this, 250, TimeUnit.MILLISECONDS, this);
        }
    }
}
