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

package com.trollworks.gcs.app;

import com.trollworks.gcs.io.UrlUtils;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.help.UpdateSystemLibraryCommand;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.utility.task.Tasks;

import java.awt.Desktop;
import java.awt.EventQueue;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.URI;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.StringTokenizer;
import java.util.concurrent.TimeUnit;
import javax.swing.JOptionPane;

/** Provides a background check for updates. */
public class UpdateChecker implements Runnable {
    private static final String  MODULE                  = "Updates";
    private static final String  LAST_VERSION_KEY        = "LastVersion";
    private static final String  LAST_LIBRARY_COMMIT_KEY = "LastLibraryCommit";
    private static       String  APP_RESULT;
    private static       String  DATA_RESULT;
    private static       boolean NEW_APP_VERSION_AVAILABLE;
    private static       boolean NEW_DATA_VERSION_AVAILABLE;
    private              Mode    mMode;

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

    private static synchronized void setAppResult(String result, boolean available) {
        APP_RESULT = result;
        NEW_APP_VERSION_AVAILABLE = available;
    }

    /** @return Whether a new data version is available. */
    public static synchronized boolean isNewDataVersionAvailable() {
        return NEW_DATA_VERSION_AVAILABLE;
    }

    /** @return The result of the new data check. */
    public static synchronized String getDataResult() {
        return DATA_RESULT != null ? DATA_RESULT : I18n.Text("Checking for GCS Master Library updates…");
    }

    private static synchronized void setDataResult(String result, boolean available) {
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
            setAppResult(null, false);
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
            setAlreadyHaveMostRecentApp();
        } else {
            long versionAvailable = GCS.VERSION;
            try {
                try (BufferedReader in = new BufferedReader(new InputStreamReader(UrlUtils.setupConnection(new URL(GCS.WEB_SITE + "/versions.txt")).getInputStream(), StandardCharsets.UTF_8))) {
                    String line = in.readLine();
                    while (line != null) {
                        StringTokenizer tokenizer = new StringTokenizer(line, "\t");
                        if (tokenizer.hasMoreTokens()) {
                            if ("gcs".equalsIgnoreCase(tokenizer.nextToken()) && tokenizer.hasMoreTokens()) {
                                String token   = tokenizer.nextToken();
                                long   version = Version.extract(token, 0);
                                if (version > versionAvailable) {
                                    versionAvailable = version;
                                }
                            }
                        }
                        line = in.readLine();
                    }
                }
            } catch (Exception exception) {
                // Don't care
            }
            if (versionAvailable > GCS.VERSION) {
                Preferences prefs = Preferences.getInstance();
                setAppResult(I18n.Text("A new version of GCS is available"), true);
                if (versionAvailable > prefs.getLongValue(MODULE, LAST_VERSION_KEY, GCS.VERSION)) {
                    prefs.setValue(MODULE, LAST_VERSION_KEY, versionAvailable);
                    prefs.save();
                    mMode = Mode.NOTIFY;
                }
            } else {
                setAlreadyHaveMostRecentApp();
            }
        }
    }

    private void checkForLibraryUpdates() {
        String latest = Library.getLatestCommit();
        if (latest.isBlank()) {
            setDataResult(I18n.Text("Unable to access GitHub to check GCS Master Library version"), false);
        } else {
            String recorded = Library.getRecordedCommit();
            if (latest.equals(recorded)) {
                setDataResult(I18n.Text("You have the most recent version of the GCS Master Library"), false);
            } else if (GCS.VERSION == 0 || GCS.VERSION >= Library.getMinimumGCSVersion()) {
                Preferences prefs = Preferences.getInstance();
                setDataResult(I18n.Text("A new version of the GCS Master Library is available"), true);
                if (!latest.equals(prefs.getStringValue(MODULE, LAST_LIBRARY_COMMIT_KEY, ""))) {
                    prefs.setValue(MODULE, LAST_LIBRARY_COMMIT_KEY, latest);
                    prefs.save();
                    mMode = Mode.NOTIFY;
                }
            } else {
                setDataResult(I18n.Text("Your version of GCS must be updated to use the latest GCS Master Library"), false);
            }
        }
    }

    private void setAlreadyHaveMostRecentApp() {
        setAppResult(I18n.Text("You have the most recent version of GCS"), false);
    }

    private void tryNotify() {
        if (UIApp.isNotificationAllowed()) {
            String update = I18n.Text("Update");
            mMode = Mode.DONE;
            if (isNewAppVersionAvailable()) {
                if (WindowUtils.showConfirmDialog(null, getAppResult(), update, JOptionPane.OK_CANCEL_OPTION, new String[]{update, I18n.Text("Ignore")}, update) == JOptionPane.OK_OPTION) {
                    goToUpdate();
                }
            } else if (isNewDataVersionAvailable()) {
                UpdateSystemLibraryCommand.askUserToUpdate();
            }
        } else {
            Tasks.scheduleOnUIThread(this, 250, TimeUnit.MILLISECONDS, this);
        }
    }
}
