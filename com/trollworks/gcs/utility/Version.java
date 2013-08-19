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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.utility.text.NumberUtils;

/** Provides basic version checking. */
public class Version {
	/**
	 * Utility routine to extract the version number from a string. Version numbers are formatted as
	 * [year].[month].[day] of release. The period (".") separator character may be substituted with
	 * the underscore ("_") character.
	 * 
	 * @param versionString The version string to parse.
	 * @return The version number.
	 */
	public static int extractVersion(String versionString) {
		String[] parts = versionString.split("[\\._]", 3); //$NON-NLS-1$
		int version = 0;
		if (parts.length > 0) {
			version = NumberUtils.getNonLocalizedInteger(parts[0], -1);
			if (version < 0 || version > 99999) {
				version = 0;
			}
			version *= 10000;
		}
		if (parts.length > 1) {
			int value = NumberUtils.getNonLocalizedInteger(parts[1], 0);
			if (value < 1 || value > 12) {
				value = 1;
			}
			version += value * 100;
		}
		if (parts.length > 2) {
			int value = NumberUtils.getNonLocalizedInteger(parts[2], 0);
			if (value < 1 || value > 31) {
				value = 1;
			}
			version += value;
		}
		return version;
	}

	/**
	 * Converts a version number returned from {@link #extractVersion(String)} back into a
	 * human-readable string.
	 * 
	 * @param version The version number.
	 * @return The human-readable version number.
	 */
	public static String getHumanReadableVersion(int version) {
		StringBuilder buffer = new StringBuilder();
		buffer.append(version / 10000);
		buffer.append('.');
		version %= 10000;
		int value = version / 100;
		if (value < 10) {
			buffer.append('0');
		}
		buffer.append(value);
		buffer.append('.');
		version %= 100;
		if (version < 10) {
			buffer.append('0');
		}
		buffer.append(version);
		return buffer.toString();
	}
}
