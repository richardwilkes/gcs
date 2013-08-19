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

import java.awt.Dimension;
import java.awt.Point;

/** Provides conversion routines from one type to another. */
public class TKConversion {
	private static final String	COMMA	= ",";	//$NON-NLS-1$

	/**
	 * Extracts the text that makes up a number from within a string.
	 * 
	 * @param text The text to extract the information from.
	 * @param allowDecimal Pass in <code>true</code> if a decimal point is allowed.
	 * @return The extracted text.
	 */
	public static String extractNumberString(String text, boolean allowDecimal) {
		int length = text.length();
		boolean hasDecimal = false;
		StringBuilder buffer = new StringBuilder(length);

		for (int i = 0; i < length; i++) {
			char ch = text.charAt(i);

			if (buffer.length() == 0) {
				if (ch == ' ' || ch == '\t' || ch == '+') {
					continue;
				}
				if (ch == '-' || allowDecimal && ch == '.' || ch >= '0' && ch <= '9') {
					if (ch == '.') {
						hasDecimal = true;
					}
					buffer.append(ch);
				} else {
					break;
				}
			} else if (allowDecimal && !hasDecimal && ch == '.' || ch >= '0' && ch <= '9') {
				if (ch == '.') {
					hasDecimal = true;
				}
				buffer.append(ch);
			} else {
				break;
			}
		}
		return buffer.toString();
	}

	/**
	 * @param dim The dimension.
	 * @return A string version of the {@link Dimension} that can be extracted using
	 *         {@link #extractDimension(String)}.
	 */
	public static String createString(Dimension dim) {
		return dim.width + COMMA + dim.height;
	}

	/**
	 * Extracts a {@link Dimension}from the string.
	 * 
	 * @param buffer The string to extract from.
	 * @return The extracted {@link Dimension}, or <code>null</code> if valid data can't be
	 *         found.
	 */
	public static Dimension extractDimension(String buffer) {
		int[] values = extractIntegers(buffer);

		if (values.length == 2) {
			return new Dimension(values[0], values[1]);
		}
		return null;
	}

	/**
	 * @param pt The point.
	 * @return A string version of the {@link Point} that can be extracted using
	 *         {@link #extractPoint(String)}.
	 */
	public static String createString(Point pt) {
		return pt.x + COMMA + pt.y;
	}

	/**
	 * Extracts a {@link Point}from the string.
	 * 
	 * @param buffer The string to extract from.
	 * @return The extracted {@link Point}, or <code>null</code> if valid data can't be found.
	 */
	public static Point extractPoint(String buffer) {
		int[] values = extractIntegers(buffer);

		if (values.length == 2) {
			return new Point(values[0], values[1]);
		}
		return null;
	}

	/**
	 * @param values The integer values.
	 * @return A string version of the int array that can be extracted using
	 *         {@link #extractIntegers(String)}.
	 */
	public static String createString(int[] values) {
		StringBuilder buffer = new StringBuilder();

		for (int i = 0; i < values.length; i++) {
			if (i != 0) {
				buffer.append(COMMA);
			}
			buffer.append(values[i]);
		}
		return buffer.toString();
	}

	/**
	 * Extracts an <code>int</code> array from the string.
	 * 
	 * @param buffer The buffer to extract from.
	 * @return An array of integers.
	 */
	public static int[] extractIntegers(String buffer) {
		if (buffer != null && buffer.length() > 0) {
			String[] buffers = buffer.split(COMMA);
			int[] values = new int[buffers.length];

			for (int i = 0; i < buffers.length; i++) {
				try {
					values[i] = Integer.parseInt(buffers[i].trim());
				} catch (Exception exception) {
					values[i] = 0;
				}
			}
			return values;
		}
		return new int[0];
	}

	/**
	 * Extracts an <code>long</code> array from the string.
	 * 
	 * @param buffer The buffer to extract from.
	 * @return An array of longs.
	 */
	public static long[] extractLongs(String buffer) {
		if (buffer != null && buffer.length() > 0) {
			String[] buffers = buffer.split(COMMA);
			long[] values = new long[buffers.length];

			for (int i = 0; i < buffers.length; i++) {
				try {
					values[i] = Long.parseLong(buffers[i].trim());
				} catch (Exception exception) {
					values[i] = 0;
				}
			}
			return values;
		}
		return new long[0];
	}

	/**
	 * Extracts an {@link Long} array from the string.
	 * 
	 * @param buffer The buffer to extract from.
	 * @return An array of {@link Long}.
	 */
	public static Long[] extractLongObjects(String buffer) {
		if (buffer != null && buffer.length() > 0) {
			String[] buffers = buffer.split(COMMA);
			Long[] values = new Long[buffers.length];

			for (int i = 0; i < buffers.length; i++) {
				try {
					values[i] = new Long(buffers[i].trim());
				} catch (Exception exception) {
					values[i] = new Long(0);
				}
			}
			return values;
		}
		return new Long[0];
	}
}
