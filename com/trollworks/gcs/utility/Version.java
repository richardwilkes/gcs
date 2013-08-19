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

/** Provides basic version checking. */
public class Version {
	/**
	 * Utility routine to extract the version number from a string. Version numbers are assumed to
	 * be formatted as [major].[minor].[bugfix].[build], where only [major] is required, although if
	 * a number is specified, all numbers up to that portion must also be specified. The period
	 * (".") separator character may be substituted with the underscore ("_") character.
	 * 
	 * @param versionString The version string to parse.
	 * @return The version number.
	 */
	public static long extractVersion(String versionString) {
		long version = 0;
		int shift = 48;
		long value = 0;

		for (char ch : versionString.toCharArray()) {
			if (ch >= '0' && ch <= '9') {
				value *= 10;
				value += ch - '0';
			} else if (ch == '.' || ch == '_') {
				if (value > 0xEFFF) {
					value = 0xEFFF;
				}
				version |= value << shift;
				value = 0;
				shift -= 16;
				if (shift < 0) {
					break;
				}
			} else {
				if (value > 0xEFFF) {
					value = 0xEFFF;
				}
				version |= value << shift;
				value = 0;
				shift -= 16;
				break;
			}
		}
		if (shift >= 0) {
			if (value > 0xEFFF) {
				value = 0xEFFF;
			}
			version |= value << shift;
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
	public static String getHumanReadableVersion(long version) {
		return getHumanReadableVersion(version, false);
	}

	/**
	 * Converts a version number returned from {@link #extractVersion(String)} back into a
	 * human-readable string.
	 * 
	 * @param version The version number.
	 * @param includeBuildNumber <code>true</code> to include all version number digits, including
	 *            the build number, even if they are zero.
	 * @return The human-readable version number.
	 */
	public static String getHumanReadableVersion(long version, boolean includeBuildNumber) {
		StringBuilder buffer = new StringBuilder();

		buffer.append(version >> 48);
		version &= 0x0000FFFFFFFFFFFFL;
		if (includeBuildNumber || (version & 0x0000FFFFFFFF0000L) > 0) {
			buffer.append('.');
			buffer.append(version >> 32);
			version &= 0x00000000FFFFFFFFL;
			if (includeBuildNumber || (version & 0x00000000FFFF0000L) > 0) {
				buffer.append('.');
				buffer.append(version >> 16);
				version &= 0x000000000000FFFFL;
				if (includeBuildNumber) {
					buffer.append('.');
					buffer.append(version);
				}
			}
		}
		return buffer.toString();
	}
}
