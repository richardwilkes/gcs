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

package com.trollworks.toolkit.collections;

/** A utility for consistent extraction of an {@link Enum} value from a text buffer. */
public class TKEnumExtractor {
	private static final String	EMPTY		= "";	//$NON-NLS-1$
	private static final String	COMMA		= ",";	//$NON-NLS-1$
	private static final String	UNDERSCORE	= "_";	//$NON-NLS-1$
	private static final String	SPACE		= " ";	//$NON-NLS-1$

	/**
	 * @param buffer The buffer to load from.
	 * @param values The possible values.
	 * @param defaultValue The default value to use in case of no match.
	 * @return The {@link Enum} representing the buffer.
	 */
	// The return value should be "Enum<?>", however, the command-line compiler seems to be choking
	// on that where this class is actually used...
	@SuppressWarnings("unchecked") public static final Enum extract(String buffer, Enum<?>[] values, Enum<?> defaultValue) {
		Enum<?> value = extract(buffer, values);

		return value != null ? value : defaultValue;
	}

	/**
	 * @param buffer The buffer to load from.
	 * @param values The possible values.
	 * @return The {@link Enum} representing the buffer, or <code>null</code> if a match could not
	 *         be found.
	 */
	// The return value should be "Enum<?>", however, the command-line compiler seems to be choking
	// on that where this class is actually used...
	@SuppressWarnings("unchecked") public static final Enum extract(String buffer, Enum<?>[] values) {
		for (Enum<?> type : values) {
			if (type.name().equalsIgnoreCase(buffer)) {
				return type;
			}
		}

		// If that failed, replace any embedded underscores in the name with
		// spaces and try again
		for (Enum<?> type : values) {
			if (type.name().replaceAll(UNDERSCORE, SPACE).equalsIgnoreCase(buffer)) {
				return type;
			}
		}

		// If that failed, replace any embedded underscores in the name with
		// commas and try again
		for (Enum<?> type : values) {
			if (type.name().replaceAll(UNDERSCORE, COMMA).equalsIgnoreCase(buffer)) {
				return type;
			}
		}

		// If that failed, remove any embedded underscores in the name and try
		// again
		for (Enum<?> type : values) {
			if (type.name().replaceAll(UNDERSCORE, EMPTY).equalsIgnoreCase(buffer)) {
				return type;
			}
		}

		return null;
	}
}
