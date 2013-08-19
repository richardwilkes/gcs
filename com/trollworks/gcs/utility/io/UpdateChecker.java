/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility.io;

import com.trollworks.gcs.app.App;
import com.trollworks.gcs.app.Main;
import com.trollworks.gcs.utility.DelayedTask;
import com.trollworks.gcs.utility.Version;
import com.trollworks.gcs.widgets.WindowUtils;

import java.awt.EventQueue;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.URL;
import java.text.MessageFormat;
import java.util.StringTokenizer;

import javax.swing.JOptionPane;

/** Provides a background check for updates. */
public class UpdateChecker implements Runnable {
	private static String		MSG_CHECKING;
	private static String		MSG_UP_TO_DATE;
	private static String		MSG_OUT_OF_DATE;
	private static String		MSG_UPDATE_TITLE;
	private static String		MSG_IGNORE_TITLE;
	private static final String	MODULE					= "Updates";			//$NON-NLS-1$
	private static final String	LAST_VERSION_KEY		= "LastVersionSeen";	//$NON-NLS-1$
	private static boolean		NEW_VERSION_AVAILABLE	= false;
	private static String		RESULT;
	private static String		UPDATE_URL;
	private String				mProductKey;
	private String				mCheckURL;
	private int					mMode;

	static {
		LocalizedMessages.initialize(UpdateChecker.class);
		RESULT = MSG_CHECKING;
	}

	/**
	 * Initiates a check for updates.
	 * 
	 * @param productKey The product key to check for.
	 * @param checkURL The URL to use for checking whether a new version is available.
	 * @param updateURL The URL to use when going to the update site.
	 */
	public static void check(String productKey, String checkURL, String updateURL) {
		Thread thread = new Thread(new UpdateChecker(productKey, checkURL), UpdateChecker.class.getSimpleName());

		UPDATE_URL = updateURL;
		thread.setPriority(Thread.NORM_PRIORITY);
		thread.setDaemon(true);
		thread.start();
	}

	/** @return Whether a new version is available. */
	public static boolean isNewVersionAvailable() {
		return NEW_VERSION_AVAILABLE;
	}

	/** @return The result. */
	public static String getResult() {
		return RESULT;
	}

	/** Go to the update location on the web, if a new version is available. */
	public static void goToUpdate() {
		if (NEW_VERSION_AVAILABLE) {
			Browser.getPreferredBrowser().openURL(UPDATE_URL);
		}
	}

	private UpdateChecker(String productKey, String checkURL) {
		mProductKey = productKey;
		mCheckURL = checkURL;
	}

	public void run() {
		if (mMode == 0) {
			long currentVersion = Version.extractVersion(Main.getVersion());
			long versionAvailable = currentVersion;

			mMode = 2;
			try {
				URL url = new URL(mCheckURL);
				BufferedReader in = new BufferedReader(new InputStreamReader(url.openStream()));
				String line = in.readLine();

				while (line != null) {
					StringTokenizer tokenizer = new StringTokenizer(line, "\t"); //$NON-NLS-1$

					if (tokenizer.hasMoreTokens()) {
						try {
							if (tokenizer.nextToken().equalsIgnoreCase(mProductKey)) {
								String token = tokenizer.nextToken();
								long version = Version.extractVersion(token);

								if (version > versionAvailable) {
									versionAvailable = version;
								}
							}
						} catch (Exception exception) {
							// Don't care
						}
					}
					line = in.readLine();
				}
			} catch (Exception exception) {
				// Don't care
			}

			if (versionAvailable > currentVersion) {
				Preferences prefs = Preferences.getInstance();
				String humanReadableVersion = Version.getHumanReadableVersion(versionAvailable, true);

				NEW_VERSION_AVAILABLE = true;
				RESULT = MessageFormat.format(MSG_OUT_OF_DATE, humanReadableVersion);
				if (versionAvailable > Version.extractVersion(prefs.getStringValue(MODULE, LAST_VERSION_KEY, Main.getVersion()))) {
					prefs.setValue(MODULE, LAST_VERSION_KEY, humanReadableVersion);
					prefs.save();
					mMode = 1;
					EventQueue.invokeLater(this);
					return;
				}
			} else {
				RESULT = MSG_UP_TO_DATE;
			}
		} else if (mMode == 1) {
			if (App.INSTANCE.isNotificationAllowed()) {
				String result = getResult();

				mMode = 2;
				if (WindowUtils.showConfirmDialog(null, result, MSG_UPDATE_TITLE, JOptionPane.OK_CANCEL_OPTION, new String[] { MSG_UPDATE_TITLE, MSG_IGNORE_TITLE }, MSG_UPDATE_TITLE) == JOptionPane.OK_OPTION) {
					goToUpdate();
				}
			} else {
				DelayedTask.schedule(this, 250);
			}
		}
	}
}
