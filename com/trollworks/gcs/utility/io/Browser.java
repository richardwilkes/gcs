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

import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.Debug;
import com.trollworks.gcs.widgets.WindowUtils;

import java.io.File;
import java.net.MalformedURLException;
import java.net.URL;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** Represents a browser. */
public class Browser {
	private static String		MSG_BAD_CMD_PATH;
	private static final String	ESCAPE_MARKER		= "@";					//$NON-NLS-1$
	private static final String	ESCAPED_SEPARATOR	= "@2";				//$NON-NLS-1$
	private static final String	ESCAPED_ESCAPE		= "@1";				//$NON-NLS-1$
	private static final String	SEPARATOR			= ":";					//$NON-NLS-1$
	private static final String	FIREFOX_TITLE		= "Firefox";			//$NON-NLS-1$
	private static final String	MODULE				= "Browser";			//$NON-NLS-1$
	private static final int	MODULE_VERSION		= 1;
	private static final String	BROWSER_KEY			= "PreferredBrowser";	//$NON-NLS-1$
	private static final String	URL_PARAM			= "%URL%";				//$NON-NLS-1$
	private String				mTitle;
	private String				mCommandPath;
	private String[]			mArguments;
	private boolean				mIsEditable;

	static {
		LocalizedMessages.initialize(Browser.class);
		Preferences.getInstance().resetIfVersionMisMatch(MODULE, MODULE_VERSION);
	}

	/** @return The default browser for the current platform. */
	public static final Browser getDefaultBrowser() {
		if (Platform.isMacintosh()) {
			return createMacSystemBrowser();
		} else if (Platform.isWindows()) {
			return createExplorerBrowser();
		}
		return createMozillaBrowser();
	}

	/** @return An array of the standard browsers for the current platform. */
	public static Browser[] getStandardBrowsers() {
		if (Platform.isMacintosh()) {
			return new Browser[] { createFirefoxBrowser(), createSafariBrowser(), createMacSystemBrowser() };
		} else if (Platform.isWindows()) {
			return new Browser[] { createExplorerBrowser(), createFirefoxBrowser() };
		}
		return new Browser[] { createFirefoxBrowser(), createMozillaBrowser(), createNetscapeBrowser() };
	}

	/** @return The standard BBEdit browser object. */
	public static Browser createMacSystemBrowser() {
		return new Browser("Default", "open", new String[] { URL_PARAM }, false); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/** @return The standard Firefox browser object. */
	public static Browser createFirefoxBrowser() {
		if (Platform.isMacintosh()) {
			return new Browser(FIREFOX_TITLE, "open", new String[] { "-b", "org.mozilla.firefox", URL_PARAM }, false); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
		} else if (Platform.isWindows()) {
			return new Browser(FIREFOX_TITLE, "C:\\Program Files\\Mozilla Firefox\\firefox.exe", new String[] { URL_PARAM }, false); //$NON-NLS-1$
		}
		return new Browser(FIREFOX_TITLE, "firefox", new String[] { URL_PARAM }, false); //$NON-NLS-1$
	}

	/** @return The standard Safari browser object. */
	public static Browser createSafariBrowser() {
		return new Browser("Safari", "open", new String[] { "-b", "com.apple.Safari", URL_PARAM }, false); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
	}

	/** @return The standard Mozilla browser object. */
	public static Browser createMozillaBrowser() {
		return new Browser("Mozilla", "mozilla", new String[] { URL_PARAM }, false); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/** @return The standard Netscape browser object. */
	public static Browser createNetscapeBrowser() {
		return new Browser("Netscape", "netscape", new String[] { URL_PARAM }, false); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/** @return The standard Explorer browser object. */
	public static Browser createExplorerBrowser() {
		return new Browser("Explorer", "explorer.exe", new String[] { URL_PARAM }, false); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * Creates a browser object from a buffer that had been previously created with a call to
	 * {@link #toEncodedString()}.
	 * 
	 * @param buffer The buffer to create the browser from.
	 * @return The newly created browser, or <code>null</code> if the encoded buffer was invalid.
	 */
	public static Browser createBrowserFromEncodedString(String buffer) {
		if (buffer != null) {
			StringTokenizer tokenizer = new StringTokenizer(buffer, SEPARATOR);

			if (tokenizer.hasMoreTokens()) {
				String title = decode(tokenizer.nextToken());
				String commandPath = ""; //$NON-NLS-1$
				ArrayList<String> args = new ArrayList<String>();

				if (tokenizer.hasMoreTokens()) {
					commandPath = decode(tokenizer.nextToken());
					while (tokenizer.hasMoreTokens()) {
						args.add(decode(tokenizer.nextToken()));
					}
				}
				return new Browser(title, commandPath, args.toArray(new String[0]), true);
			}
		}
		return null;
	}

	private static String decode(String buffer) {
		return buffer.replaceAll(ESCAPED_SEPARATOR, SEPARATOR).replaceAll(ESCAPED_ESCAPE, ESCAPE_MARKER);
	}

	/** @return The preferred browser. */
	public static Browser getPreferredBrowser() {
		String buffer = Preferences.getInstance().getStringValue(MODULE, BROWSER_KEY);
		Browser browser = buffer != null ? createBrowserFromEncodedString(buffer) : null;

		if (browser == null) {
			browser = getDefaultBrowser();
		}
		return browser;
	}

	/** @param browser Sets the preferred browser. */
	public static void setPreferredBrowser(Browser browser) {
		Preferences prefs = Preferences.getInstance();

		if (browser != null) {
			prefs.setValue(MODULE, BROWSER_KEY, browser.toEncodedString());
		} else {
			prefs.removePreference(MODULE, BROWSER_KEY);
		}
	}

	/**
	 * Creates a new browser.
	 * 
	 * @param title The title to give this browser.
	 * @param commandPath The command path to use.
	 * @param arguments The arguments to pass to the browser.
	 */
	public Browser(String title, String commandPath, String[] arguments) {
		this(title, commandPath, arguments, true);
	}

	private Browser(String title, String commandPath, String[] arguments, boolean editable) {
		mTitle = title;
		mCommandPath = commandPath;
		mIsEditable = true;
		setArguments(arguments);
		mIsEditable = editable;
	}

	/**
	 * Creates a clone of an existing browser object.
	 * 
	 * @param other The browser object to clone.
	 * @param editable The editable state for the clone.
	 */
	public Browser(Browser other, boolean editable) {
		this(other.mTitle, other.mCommandPath, other.mArguments, editable);
	}

	/**
	 * Opens the file in the browser.
	 * 
	 * @param file The file to open.
	 */
	public void openFile(String file) {
		openFile(new File(file));
	}

	/**
	 * Opens the file in the browser.
	 * 
	 * @param file The file to open.
	 */
	public void openFile(File file) {
		if (Platform.isMacintosh()) {
			openURL(file.getAbsolutePath());
		} else {
			try {
				openURL(file.getAbsoluteFile().toURI().toURL());
			} catch (MalformedURLException mue) {
				assert false : Debug.throwableToString(mue);
			}
		}
	}

	/**
	 * Opens the URL in the browser.
	 * 
	 * @param url The URL to open.
	 */
	public void openURL(URL url) {
		openURL(url.toExternalForm());
	}

	/**
	 * Opens the URL in the browser.
	 * 
	 * @param url The URL to open.
	 */
	public void openURL(String url) {
		String cmdPath = getCommandPath();
		File cmdFile = Path.isCommandPathViable(cmdPath, null);

		if (cmdFile == null) {
			WindowUtils.showError(null, MessageFormat.format(MSG_BAD_CMD_PATH, cmdPath));
			return;
		}

		try {
			String[] cmds = new String[1 + mArguments.length];

			cmds[0] = cmdFile.getAbsolutePath();
			for (int i = 0; i < mArguments.length; i++) {
				int index;

				cmds[i + 1] = mArguments[i];
				index = cmds[i + 1].indexOf(URL_PARAM);
				if (index != -1) {
					int length = cmds[i + 1].length();
					StringBuilder buffer = new StringBuilder();

					if (index > 0) {
						buffer.append(cmds[i + 1].substring(0, index));
					}
					buffer.append(url);
					index += URL_PARAM.length();
					if (index < length) {
						buffer.append(cmds[i + 1].substring(index));
					}
					cmds[i + 1] = buffer.toString();
				}
			}

			Runtime.getRuntime().exec(cmds);
		} catch (Exception ex) {
			assert false : Debug.throwableToString(ex);
		}
	}

	/** @return The title. */
	public String getTitle() {
		return mTitle;
	}

	/** @param title The title. */
	public void setTitle(String title) {
		if (mIsEditable) {
			mTitle = title;
		}
	}

	@Override public String toString() {
		return mTitle;
	}

	/** @return The command path. */
	public String getCommandPath() {
		return mCommandPath;
	}

	/** @param path The command path. */
	public void setCommandPath(String path) {
		if (mIsEditable) {
			mCommandPath = path;
		}
	}

	/** @return The arguments. */
	public String[] getArguments() {
		return mArguments;
	}

	/** @param arguments The arguments. */
	public void setArguments(String[] arguments) {
		if (mIsEditable) {
			int length = arguments != null ? arguments.length : 0;

			mArguments = new String[length];
			if (length > 0) {
				System.arraycopy(arguments, 0, mArguments, 0, length);
			}
		}
	}

	/** @return <code>true</code> if this browser object is editable. */
	public boolean isEditable() {
		return mIsEditable;
	}

	/** @return An encoded version of this browser object. */
	public String toEncodedString() {
		StringBuilder buffer = new StringBuilder();

		buffer.append(encode(mTitle));
		buffer.append(SEPARATOR);
		buffer.append(encode(mCommandPath));
		for (String element : mArguments) {
			buffer.append(SEPARATOR);
			buffer.append(encode(element));
		}
		return buffer.toString();
	}

	private String encode(String buffer) {
		return buffer.replaceAll(ESCAPE_MARKER, ESCAPED_ESCAPE).replaceAll(SEPARATOR, ESCAPED_SEPARATOR);
	}

	/**
	 * @param other The other browser object.
	 * @return <code>true</code> if the other browser object is equivalent to this one.
	 */
	public boolean isEquivalent(Browser other) {
		if (other == this) {
			return true;
		}
		if (mTitle.equals(other.mTitle) && mCommandPath.equals(other.mCommandPath) && mArguments.length == other.mArguments.length) {
			for (String element : mArguments) {
				if (!element.equals(element)) {
					return false;
				}
			}
			return true;
		}
		return false;
	}

	/** @return The command line for this browser. */
	public String getCommandLine() {
		StringBuilder buffer = new StringBuilder();

		quoteIfNeeded(buffer, mCommandPath);
		for (String element : mArguments) {
			buffer.append(' ');
			quoteIfNeeded(buffer, element);
		}
		return buffer.toString();
	}

	private void quoteIfNeeded(StringBuilder buffer, String arg) {
		if (arg.indexOf(' ') != -1 || arg.indexOf('\t') != -1) {
			buffer.append('"');
			buffer.append(arg);
			buffer.append('"');
		} else {
			buffer.append(arg);
		}
	}

	@Override public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof Browser) {
			Browser other = (Browser) obj;
			if (mTitle.equals(other.mTitle)) {
				if (mCommandPath.equals(other.mCommandPath)) {
					if (mArguments.length == other.mArguments.length) {
						for (int i = 0; i < mArguments.length; i++) {
							if (!mArguments[i].equals(other.mArguments[i])) {
								return false;
							}
						}
						return true;
					}
				}
			}
		}
		return false;
	}
}
