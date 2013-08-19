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

package com.trollworks.gcs.utility.text;

import java.text.NumberFormat;
import java.util.Comparator;

/**
 * A string comparator that will honor numeric values embedded in strings and treat them as numbers
 * for comparison purposes.
 */
public class NumericStringComparator implements Comparator<String> {
	/** The standard caseless, numeric-aware string comparator. */
	public static final NumericStringComparator	CASELESS_COMPARATOR	= new NumericStringComparator(true);
	/** The standard case-sensitive, numeric-aware string comparator. */
	public static final NumericStringComparator	COMPARATOR			= new NumericStringComparator(false);
	private static NumberFormat						DECIMAL_FORMATTER	= null;
	private boolean									mCaseless;

	private NumericStringComparator(boolean caseless) {
		mCaseless = caseless;
	}

	/**
	 * Convenience method for caseless numeric string comparisons.
	 * 
	 * @param s0 The first string.
	 * @param s1 The second string.
	 * @return A negative integer, zero, or a positive integer if the first argument is less than,
	 *         equal to, or greater than the second.
	 */
	public static int caselessCompareStrings(String s0, String s1) {
		return CASELESS_COMPARATOR.compare(s0, s1);
	}

	/**
	 * Convenience method for case-sensitive numeric string comparisons.
	 * 
	 * @param s0 The first string.
	 * @param s1 The second string.
	 * @return A negative integer, zero, or a positive integer if the first argument is less than,
	 *         equal to, or greater than the second.
	 */
	public static int compareStrings(String s0, String s1) {
		return COMPARATOR.compare(s0, s1);
	}

	private static final NumberFormat getDecimalFormatter() {
		if (DECIMAL_FORMATTER == null) {
			DECIMAL_FORMATTER = NumberFormat.getInstance();
			DECIMAL_FORMATTER.setMinimumFractionDigits(16);
			DECIMAL_FORMATTER.setMaximumFractionDigits(16);
		}
		return DECIMAL_FORMATTER;
	}

	private static final boolean isNumericPortion(char ch) {
		return ch == '.' || ch == ',' || ch >= '0' && ch <= '9';
	}

	public int compare(String string0, String string1) {
		if (string0 == null) {
			string0 = ""; //$NON-NLS-1$
		}
		if (string1 == null) {
			string1 = ""; //$NON-NLS-1$
		}
		NumberFormat formatter = getDecimalFormatter();
		char[] chars0 = string0.toCharArray();
		char[] chars1 = string1.toCharArray();
		int pos0 = 0;
		int pos1 = 0;
		int len0 = chars0.length;
		int len1 = chars1.length;
		int result = 0;
		int secondaryResult = 0;
		char c0;
		char c1;

		while (result == 0 && pos0 < len0 && pos1 < len1) {
			boolean normalCompare = true;

			c0 = chars0[pos0++];
			c1 = chars1[pos1++];

			if (isNumericPortion(c0) && isNumericPortion(c1)) {
				boolean foundDigit0 = false;
				boolean foundDigit1 = false;
				int count0 = 1;
				int count1 = 1;

				if (!foundDigit0 && c0 >= '0' && c0 <= '9') {
					foundDigit0 = true;
				}
				while (pos0 < len0) {
					c0 = chars0[pos0];
					if (isNumericPortion(c0)) {
						if (!foundDigit0 && c0 >= '0' && c0 <= '9') {
							foundDigit0 = true;
						}
						count0++;
						pos0++;
					} else {
						break;
					}
				}

				if (!foundDigit1 && c1 >= '0' && c1 <= '9') {
					foundDigit1 = true;
				}
				while (pos1 < len1) {
					c1 = chars1[pos1];
					if (isNumericPortion(c1)) {
						if (!foundDigit1 && c1 >= '0' && c1 <= '9') {
							foundDigit1 = true;
						}
						count1++;
						pos1++;
					} else {
						break;
					}
				}

				if (foundDigit0 && foundDigit1) {
					try {
						double val0 = formatter.parse(string0.substring(pos0 - count0, pos0)).doubleValue();
						double val1 = formatter.parse(string1.substring(pos1 - count1, pos1)).doubleValue();

						normalCompare = false;
						if (val0 > val1) {
							result = 1;
						} else if (val0 < val1) {
							result = -1;
						} else {
							result = 0;
						}
						if (result == 0 && secondaryResult == 0) {
							secondaryResult = count0 - count1;
						}
					} catch (Exception ex) {
						// Ignore
					}
				}

				if (normalCompare) {
					pos0 -= count0;
					pos1 -= count1;
					c0 = chars0[pos0++];
					c1 = chars1[pos1++];
				}
			}

			if (normalCompare) {
				if (mCaseless) {
					int c0Val = Character.isLowerCase(c0) ? Character.toUpperCase(c0) : c0;
					int c1Val = Character.isLowerCase(c1) ? Character.toUpperCase(c1) : c1;

					result = c0Val - c1Val;
				} else {
					result = c0 - c1;
				}
			}
		}

		// string without suffix comes first
		if (result == 0) {
			result = len0 - pos0 - (len1 - pos1);
		}

		// tie breaker
		if (result == 0) {
			result = secondaryResult;
		}

		if (result < 0) {
			result = -1;
		} else if (result > 0) {
			result = 1;
		}

		return result;
	}
}
