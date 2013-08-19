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

import java.io.PrintWriter;
import java.io.StringWriter;

/** Provides debugging utilities. */
public class TKDebug {
	/** Standard debug key for enabling diagnosis for load/save failures. */
	public static final String	KEY_DIAGNOSE_LOAD_SAVE	= "DEBUG_DIAGNOSE_LOAD_SAVE";	//$NON-NLS-1$

	/**
	 * @param throwable The throwable to generate a string for.
	 * @return The string form of the {@link Throwable}.
	 */
	public static String throwableToString(Throwable throwable) {
		StringWriter swriter = new StringWriter();
		PrintWriter pwriter = new PrintWriter(swriter);

		throwable.printStackTrace(pwriter);
		pwriter.flush();
		return swriter.toString();
	}

	/**
	 * Checks to see if the specified key was set to a "true" value in the system properties. If the
	 * system properties doesn't have the key set at all, then the environment is checked for the
	 * specified key instead.
	 * 
	 * @param key The key to check.
	 * @return <code>true</code> if the key is set and its value is something other than "0",
	 *         "false", "no", or "off".
	 */
	public static boolean isKeySet(String key) {
		String value = System.getProperty(key);

		if (value == null) {
			value = System.getenv(key);
		}

		return value != null && !(value.equals("0") || value.equalsIgnoreCase("false") || value.equalsIgnoreCase("off") || value.equalsIgnoreCase("no")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
	}
}
