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

package com.trollworks.toolkit.io;

import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKAppWithUI;
import com.trollworks.toolkit.utility.TKBrowser;
import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.utility.TKVersion;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKOptionDialog;

import java.awt.EventQueue;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.URL;
import java.text.MessageFormat;
import java.util.StringTokenizer;

/** Provides a background check for updates. */
public class TKUpdateChecker implements Runnable {
	private static final String	MODULE					= "Updates";			//$NON-NLS-1$
	private static final String	LAST_VERSION_KEY		= "LastVersionSeen";	//$NON-NLS-1$
	private static boolean		NEW_VERSION_AVAILABLE	= false;
	private static String		RESULT					= Msgs.CHECKING;
	private static String		UPDATE_URL;
	private String				mProductKey;
	private String				mCheckURL;
	private int					mMode;

	/**
	 * Initiates a check for updates.
	 * 
	 * @param productKey The product key to check for.
	 * @param checkURL The URL to use for checking whether a new version is available.
	 * @param updateURL The URL to use when going to the update site.
	 */
	public static void check(String productKey, String checkURL, String updateURL) {
		Thread thread = new Thread(new TKUpdateChecker(productKey, checkURL), TKUpdateChecker.class.getSimpleName());

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
			TKBrowser.getPreferredBrowser().openURL(UPDATE_URL);
		}
	}

	private TKUpdateChecker(String productKey, String checkURL) {
		mProductKey = productKey;
		mCheckURL = checkURL;
	}

	public void run() {
		if (mMode == 0) {
			long currentVersion = TKVersion.extractVersion(TKApp.getVersion());
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
								long version = TKVersion.extractVersion(token);

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
				TKPreferences prefs = TKPreferences.getInstance();
				String humanReadableVersion = TKVersion.getHumanReadableVersion(versionAvailable, true);

				NEW_VERSION_AVAILABLE = true;
				RESULT = MessageFormat.format(Msgs.OUT_OF_DATE, humanReadableVersion);
				if (versionAvailable > TKVersion.extractVersion(prefs.getStringValue(MODULE, LAST_VERSION_KEY, TKApp.getVersion()))) {
					prefs.setValue(MODULE, LAST_VERSION_KEY, humanReadableVersion);
					prefs.save();
					mMode = 1;
					EventQueue.invokeLater(this);
					return;
				}
			} else {
				RESULT = Msgs.UP_TO_DATE;
			}
		} else if (mMode == 1) {
			if (TKAppWithUI.getInstanceWithUI().isNotificationAllowed()) {
				String result = getResult();

				mMode = 2;
				if (TKOptionDialog.modal(null, TKImage.getMessageIcon(), result, TKOptionDialog.TYPE_OK_CANCEL, result, null, new String[] { Msgs.UPDATE_TITLE, Msgs.IGNORE_TITLE, "" }) == TKDialog.OK) { //$NON-NLS-1$
					goToUpdate();
				}
			} else {
				TKTimerTask.schedule(this, 250);
			}
		}
	}
}
