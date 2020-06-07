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
import com.trollworks.gcs.menu.help.UpdateMasterLibraryCommand;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.MarkdownDocument;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.task.Tasks;

import java.awt.Desktop;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.net.URI;
import java.net.URL;
import java.util.concurrent.TimeUnit;
import javax.swing.JOptionPane;
import javax.swing.JScrollPane;
import javax.swing.JTextPane;

/** Provides a background check for updates. */
public class UpdateChecker implements Runnable {
    private static String  APP_RESULT;
    private static String  APP_RELEASE_NOTES;
    private static String  DATA_RESULT;
    private static boolean NEW_APP_VERSION_AVAILABLE;
    private static boolean NEW_DATA_VERSION_AVAILABLE;
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

    /** @return Whether a new data version is available. */
    public static synchronized boolean isNewDataVersionAvailable() {
        return NEW_DATA_VERSION_AVAILABLE;
    }

    /** @return The result of the new data check. */
    public static synchronized String getDataResult() {
        return DATA_RESULT != null ? DATA_RESULT : I18n.Text("Checking for Master Library updates…");
    }

    public static synchronized void setDataResult(String result, boolean available) {
        DATA_RESULT = result;
        NEW_DATA_VERSION_AVAILABLE = available;
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
            setDataResult(null, false);
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
        if (GCS.VERSION == 0) {
            // Development version. Bail.
            setAppResult(I18n.Text("Development versions don't look for GCS updates"), null, false);
        } else {
            long   versionAvailable = GCS.VERSION;
            String releaseNotes     = "";
            try {
                JsonMap m   = Json.asMap(Json.parse(new URL("https://api.github.com/repos/richardwilkes/gcs/releases/latest")), false);
                String  tag = m.getString("tag_name", false);
                if (tag.startsWith("v")) {
                    long version = Version.extract(tag.substring(1), 0);
                    if (version > versionAvailable) {
                        versionAvailable = version;
                        releaseNotes = m.getString("body", false);
                    }
                }
            } catch (Exception exception) {
                Log.error(exception);
            }
            if (versionAvailable > GCS.VERSION) {
                Preferences prefs = Preferences.getInstance();
                setAppResult(String.format(I18n.Text("GCS v%s is available!"), Version.toString(versionAvailable, false)), releaseNotes, true);
                if (versionAvailable > prefs.getLastGCSVersion()) {
                    prefs.setLastGCSVersion(versionAvailable);
                    prefs.save();
                    mMode = Mode.NOTIFY;
                }
            } else {
                setAppResult(I18n.Text("You have the most recent version of GCS"), null, false);
            }
        }
    }

    private void checkForLibraryUpdates() {
        String latest = Library.getLatestCommit();
        if (latest.isBlank()) {
            setDataResult(I18n.Text("Unable to access GitHub to check the Master Library version"), false);
        } else {
            String recorded = Library.getRecordedCommit();
            if (latest.equals(recorded)) {
                setDataResult(I18n.Text("You have the most recent version of the Master Library"), false);
            } else if (GCS.VERSION == 0 || GCS.VERSION >= Library.getMinimumGCSVersion()) {
                Preferences prefs = Preferences.getInstance();
                setDataResult(I18n.Text("A new version of the Master Library is available"), true);
                if (!latest.equals(prefs.getLatestLibraryCommit())) {
                    prefs.setLatestLibraryCommit(latest);
                    prefs.save();
                    mMode = Mode.NOTIFY;
                }
            } else {
                setDataResult(I18n.Text("A newer version of GCS is required to use the latest Master Library"), false);
            }
        }
    }

    private void tryNotify() {
        if (GCS.isNotificationAllowed()) {
            String update = I18n.Text("Update");
            mMode = Mode.DONE;
            if (isNewAppVersionAvailable()) {
                JTextPane markdown = new JTextPane(new MarkdownDocument(getAppReleaseNotes()));
                Dimension size = markdown.getPreferredSize();
                JScrollPane scroller = new JScrollPane(markdown);
                int maxWidth = Math.min(600, WindowUtils.getMaximumWindowBounds().width * 3 / 2);
                if (size.width > maxWidth) {
                    markdown.setSize(new Dimension(maxWidth, Short.MAX_VALUE));
                    size = markdown.getPreferredSize();
                    size.width = maxWidth;
                    markdown.setPreferredSize(size);
                }
                if (WindowUtils.showOptionDialog(null, scroller, getAppResult(), true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.QUESTION_MESSAGE, null, new String[]{update, I18n.Text("Ignore")}, update) == JOptionPane.OK_OPTION) {
                    goToUpdate();
                }
            } else if (isNewDataVersionAvailable()) {
                UpdateMasterLibraryCommand.askUserToUpdate();
            }
        } else {
            Tasks.scheduleOnUIThread(this, 250, TimeUnit.MILLISECONDS, this);
        }
    }
}
