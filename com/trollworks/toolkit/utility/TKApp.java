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

package com.trollworks.toolkit.utility;

import java.text.MessageFormat;

/** Provides common application information, such as name and version. */
public class TKApp {
	private static final String	UNKNOWN				= "?";		//$NON-NLS-1$
	private static String		APP_NAME			= UNKNOWN;
	private static String		APP_VERSION			= "1.0";	//$NON-NLS-1$
	private static String		APP_COPYRIGHT_YEARS	= UNKNOWN;
	private static String		APP_COPYRIGHT_OWNER	= UNKNOWN;
	/** The one and only instance of the application object. */
	protected static TKApp		INSTANCE;

	/** Creates a new {@link TKApp}. */
	public TKApp() {
		INSTANCE = this;
	}

	/** @return The last instance created. */
	public static TKApp getInstance() {
		return INSTANCE;
	}

	/**
	 * @param useRealCopyrightSymbol Whether or not a real copyright symbol should be used.
	 * @return A version banner for this application.
	 */
	public static String getVersionBanner(boolean useRealCopyrightSymbol) {
		return MessageFormat.format(Msgs.LONG_VERSION_FORMAT, APP_NAME, APP_VERSION, getCopyrightBanner(useRealCopyrightSymbol));
	}

	/** @return A version banner for this application. */
	public static String getShortVersionBanner() {
		return MessageFormat.format(Msgs.SHORT_VERSION_FORMAT, APP_NAME, APP_VERSION);
	}

	/**
	 * @param useRealCopyrightSymbol Whether or not a real copyright symbol should be used.
	 * @return A copyright banner for this application.
	 */
	public static String getCopyrightBanner(boolean useRealCopyrightSymbol) {
		String banner = MessageFormat.format(Msgs.COPYRIGHT, APP_COPYRIGHT_YEARS, APP_COPYRIGHT_OWNER);

		if (useRealCopyrightSymbol) {
			banner = banner.replaceAll("\\(c\\)", "\u00A9"); //$NON-NLS-1$ //$NON-NLS-2$
		}
		return banner;
	}

	/** @return The name of this application. */
	public static String getName() {
		return APP_NAME;
	}

	/** @param name The name of this application. */
	public static void setName(String name) {
		APP_NAME = name;
	}

	/** @return The version of this application. */
	public static String getVersion() {
		return APP_VERSION;
	}

	/** @param version The version of this application. */
	public static void setVersion(String version) {
		APP_VERSION = version;
	}

	/** @return The copyright years for this application. */
	public static String getCopyrightYears() {
		return APP_COPYRIGHT_YEARS;
	}

	/** @param years The copyright years for this application. */
	public static void setCopyrightYears(String years) {
		APP_COPYRIGHT_YEARS = years;
	}

	/** @return The copyright owner for this application. */
	public static String getCopyrightOwner() {
		return APP_COPYRIGHT_OWNER;
	}

	/** @param owner The copyright owner for this application. */
	public static void setCopyrightOwner(String owner) {
		APP_COPYRIGHT_OWNER = owner;
	}
}
