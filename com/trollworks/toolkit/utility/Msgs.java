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

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKApp}. */
	public static String	SHORT_VERSION_FORMAT;
	/** Used by {@link TKApp}. */
	public static String	LONG_VERSION_FORMAT;
	/** Used by {@link TKApp}. */
	public static String	COPYRIGHT;

	/** Used by {@link TKAppWithUI}. */
	public static String	CONFIRM_QUIT;

	/** Used by {@link TKBrowser}. */
	public static String	BAD_CMD_PATH;

	/** Used by {@link TKTiming}. */
	public static String	ONE_SECOND;
	/** Used by {@link TKTiming}. */
	public static String	SECONDS_FORMAT;

	/** Used by {@link TKGraphics}. */
	public static String	HEADLESS;

	/** Used by {@link TKKeystroke}. */
	public static String	COMMAND;
	/** Used by {@link TKKeystroke}. */
	public static String	CONTROL;
	/** Used by {@link TKKeystroke}. */
	public static String	ALT;
	/** Used by {@link TKKeystroke}. */
	public static String	META;
	/** Used by {@link TKKeystroke}. */
	public static String	SHIFT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
