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

import java.text.DateFormat;
import java.text.DecimalFormat;
import java.text.MessageFormat;
import java.text.NumberFormat;
import java.util.Date;

/** Provides number parsing and formatting. */
public class TKNumberUtils {
	private static final String			PLUS	= "+";			//$NON-NLS-1$
	private static final String			EMPTY	= "";			//$NON-NLS-1$
	private static final DecimalFormat	NUMBER_FORMAT;
	private static final DecimalFormat	NO_COMMA_NUMBER_FORMAT;

	static {
		NUMBER_FORMAT = (DecimalFormat) NumberFormat.getNumberInstance();
		NUMBER_FORMAT.setMaximumFractionDigits(13);

		NO_COMMA_NUMBER_FORMAT = (DecimalFormat) NumberFormat.getNumberInstance();
		NO_COMMA_NUMBER_FORMAT.setGroupingUsed(false);
		NO_COMMA_NUMBER_FORMAT.setMaximumFractionDigits(13);
	}

	/**
	 * @param value The value to format.
	 * @return The formatted string.
	 */
	public static String format(long value) {
		return NUMBER_FORMAT.format(value);
	}

	/**
	 * @param value The value to format.
	 * @return The formatted string.
	 */
	public static String format(double value) {
		return NUMBER_FORMAT.format(value);
	}

	/**
	 * @param value The value to format.
	 * @return The formatted string.
	 */
	public static String formatNoComma(double value) {
		return NO_COMMA_NUMBER_FORMAT.format(value);
	}

	/**
	 * @param value The value to format.
	 * @param forceSign Pass in <code>true</code> to force a "+" sign at the beginning of positive
	 *            values.
	 * @return The formatted string.
	 */
	public static String format(long value, boolean forceSign) {
		return (forceSign && value >= 0 ? PLUS : EMPTY) + NUMBER_FORMAT.format(value);
	}

	/**
	 * @param value The value to format.
	 * @param forceSign Pass in <code>true</code> to force a "+" sign at the beginning of positive
	 *            values.
	 * @return The formatted string.
	 */
	public static String format(double value, boolean forceSign) {
		return (forceSign && value >= 0.0 ? PLUS : EMPTY) + NUMBER_FORMAT.format(value);
	}

	/**
	 * @param buffer The string to convert.
	 * @return The boolean value of the string.
	 */
	public static boolean getBoolean(String buffer) {
		buffer = strip(buffer);
		return "true".equalsIgnoreCase(buffer) || "yes".equalsIgnoreCase(buffer) || "on".equalsIgnoreCase(buffer) || "1".equals(buffer); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
	}

	private static String strip(String buffer) {
		if (buffer != null) {
			buffer = buffer.trim();
			return buffer.startsWith(PLUS) ? buffer.substring(1) : buffer;
		}
		return null;
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The short value of a string.
	 */
	public static short getShort(String buffer, short defValue) {
		try {
			return NUMBER_FORMAT.parse(strip(buffer)).shortValue();
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The short value of a string.
	 */
	public static short getNonLocalizedShort(String buffer, short defValue) {
		try {
			return Short.parseShort(strip(buffer));
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The integer value of a string.
	 */
	public static int getInteger(String buffer, int defValue) {
		try {
			return NUMBER_FORMAT.parse(strip(buffer)).intValue();
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The integer value of a string.
	 */
	public static int getNonLocalizedInteger(String buffer, int defValue) {
		try {
			return Integer.parseInt(strip(buffer));
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The long value of a string.
	 */
	public static long getLong(String buffer, long defValue) {
		try {
			return NUMBER_FORMAT.parse(strip(buffer)).longValue();
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The long value of a string.
	 */
	public static long getNonLocalizedLong(String buffer, long defValue) {
		try {
			return Long.parseLong(strip(buffer));
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The float value of a string.
	 */
	public static float getFloat(String buffer, float defValue) {
		try {
			return NUMBER_FORMAT.parse(strip(buffer)).floatValue();
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The float value of a string.
	 */
	public static float getNonLocalizedFloat(String buffer, float defValue) {
		try {
			return Float.parseFloat(strip(buffer));
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The double value of a string.
	 */
	public static double getDouble(String buffer, double defValue) {
		try {
			return NUMBER_FORMAT.parse(strip(buffer)).doubleValue();
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @param defValue The default value to use if extraction fails.
	 * @return The double value of a string.
	 */
	public static double getNonLocalizedDouble(String buffer, double defValue) {
		try {
			return Double.parseDouble(strip(buffer));
		} catch (Exception exception) {
			return defValue;
		}
	}

	/**
	 * @param buffer The string to convert.
	 * @return The number of milliseconds since midnight, January 1, 1970.
	 */
	public static long getDate(String buffer) {
		if (buffer != null) {
			buffer = buffer.trim();

			for (int i = DateFormat.FULL; i <= DateFormat.SHORT; i++) {
				try {
					return DateFormat.getDateInstance(i).parse(buffer).getTime();
				} catch (Exception exception) {
					// Ignore
				}
			}
		}
		return System.currentTimeMillis();
	}

	/**
	 * @param buffer The string to convert.
	 * @return The number of milliseconds since midnight, January 1, 1970.
	 */
	public static long getDateTime(String buffer) {
		if (buffer != null) {
			buffer = buffer.trim();

			for (int i = DateFormat.FULL; i <= DateFormat.SHORT; i++) {
				for (int j = DateFormat.FULL; j <= DateFormat.SHORT; j++) {
					try {
						return DateFormat.getDateTimeInstance(i, j).parse(buffer).getTime();
					} catch (Exception exception) {
						// Ignore
					}
				}
			}
		}
		return System.currentTimeMillis();
	}

	/**
	 * @param dateTime The date time value to format.
	 * @param format The format string to use. The first argument will be the time and the second
	 *            will be the date.
	 * @return The formatted date and time.
	 */
	public static String formatDateTime(long dateTime, String format) {
		Date date = new Date(dateTime);

		return MessageFormat.format(format, DateFormat.getTimeInstance(DateFormat.SHORT).format(date), DateFormat.getDateInstance(DateFormat.MEDIUM).format(date));
	}

	/**
	 * @param inches The height, in inches.
	 * @return The formatted height, in feet/inches format, as appropriate.
	 */
	public static String formatHeight(int inches) {
		int feet = inches / 12;

		inches %= 12;

		if (feet > 0) {
			if (inches > 0) {
				return format(feet) + "' " + format(inches) + "\""; //$NON-NLS-1$ //$NON-NLS-2$
			}
			return format(feet) + "'"; //$NON-NLS-1$
		}
		return format(inches) + "\""; //$NON-NLS-1$
	}

	/**
	 * @param buffer The formatted height string, as produced by {@link #formatHeight(int)}.
	 * @return The number of inches it represents.
	 */
	public static int getHeight(String buffer) {
		if (buffer == null) {
			return 0;
		}

		int feetMark = buffer.indexOf("'"); //$NON-NLS-1$
		int inchesMark = buffer.indexOf('"');

		if (feetMark == -1 && inchesMark == -1) {
			return getInteger(buffer, 0);
		}

		if (feetMark == -1) {
			return getInteger(buffer.substring(0, inchesMark), 0);
		}

		int inches = getInteger(buffer.substring(inchesMark != -1 && feetMark > inchesMark ? inchesMark + 1 : 0, feetMark), 0) * 12;
		if (inchesMark != -1) {
			inches += TKNumberUtils.getDouble(buffer.substring(feetMark < inchesMark ? feetMark + 1 : 0, inchesMark), 0);
		}
		return inches;
	}
}
