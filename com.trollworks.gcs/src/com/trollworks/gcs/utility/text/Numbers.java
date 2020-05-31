/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.text;

import java.text.DateFormat;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.text.NumberFormat;
import java.util.regex.Pattern;

/** Various number utilities. */
public class Numbers {
    public static final  String        YES                               = "yes";
    public static final  String        NO                                = "no";
    public static final  String        LOCALIZED_DECIMAL_SEPARATOR       = Character.toString(DecimalFormatSymbols.getInstance().getDecimalSeparator());
    private static final String        SAFE_LOCALIZED_GROUPING_SEPARATOR = Pattern.quote(Character.toString(DecimalFormatSymbols.getInstance().getGroupingSeparator()));
    private static final DecimalFormat NUMBER_FORMAT;
    private static final DecimalFormat NUMBER_PLUS_FORMAT;

    static {
        NUMBER_FORMAT = (DecimalFormat) NumberFormat.getNumberInstance();
        NUMBER_FORMAT.setMaximumFractionDigits(5);

        NUMBER_PLUS_FORMAT = (DecimalFormat) NUMBER_FORMAT.clone();
        NUMBER_PLUS_FORMAT.setPositivePrefix("+");
    }

    /**
     * @param buffer The text to process.
     * @return {@code true} if the buffer contains a 'true' value.
     */
    public static boolean extractBoolean(String buffer) {
        buffer = normalizeNumber(buffer, false);
        return "true".equalsIgnoreCase(buffer) || YES.equalsIgnoreCase(buffer) || "on".equalsIgnoreCase(buffer) || "1".equals(buffer);
    }

    /**
     * Extracts a value from the specified buffer. In addition to typical input, this method can
     * also handle some suffixes:
     * <ul>
     * <li>'b', 'B', 'g' or 'G' &mdash; multiplies the value by one billion</li>
     * <li>'m' or 'M' &mdash; multiplies the value by one million</li>
     * <li>'t', 'T', 'k' or 'K' &mdash; multiplies the value by one thousand</li>
     * </ul>
     *
     * @param buffer    The text to process.
     * @param def       The default value to return, if the buffer cannot be parsed.
     * @param localized {@code true} if the text was localized.
     * @return The value.
     */
    public static final int extractInteger(String buffer, int def, boolean localized) {
        buffer = normalizeNumber(buffer, localized);
        if (hasDecimalSeparator(buffer, localized)) {
            return (int) extractDouble(buffer, def, localized);
        }
        int multiplier = 1;
        if (hasBillionsSuffix(buffer)) {
            multiplier = 1000000000;
            buffer = removeSuffix(buffer);
        } else if (hasMillionsSuffix(buffer)) {
            multiplier = 1000000;
            buffer = removeSuffix(buffer);
        } else if (hasThousandsSuffix(buffer)) {
            multiplier = 1000;
            buffer = removeSuffix(buffer);
        }
        int max = Integer.MAX_VALUE / multiplier;
        int min = Integer.MIN_VALUE / multiplier;
        try {
            int value = Integer.parseInt(buffer);
            if (value > max) {
                value = max;
            } else if (value < min) {
                value = min;
            }
            return value * multiplier;
        } catch (Exception exception) {
            return def;
        }
    }

    /**
     * Extracts a value from the specified buffer. In addition to typical input, this method can
     * also handle some suffixes:
     * <ul>
     * <li>'b', 'B', 'g' or 'G' &mdash; multiplies the value by one billion</li>
     * <li>'m' or 'M' &mdash; multiplies the value by one million</li>
     * <li>'t', 'T', 'k' or 'K' &mdash; multiplies the value by one thousand</li>
     * </ul>
     *
     * @param buffer    The text to process.
     * @param def       The default value to return, if the buffer cannot be parsed.
     * @param min       The minimum value to return.
     * @param max       The maximum value to return.
     * @param localized {@code true} if the text was localized.
     * @return The value.
     */
    public static final int extractInteger(String buffer, int def, int min, int max, boolean localized) {
        return Math.min(Math.max(extractInteger(buffer, def, localized), min), max);
    }

    /**
     * Extracts a value from the specified buffer. In addition to typical input, this method can
     * also handle some suffixes:
     * <ul>
     * <li>'b', 'B', 'g' or 'G' &mdash; multiplies the value by one billion</li>
     * <li>'m' or 'M' &mdash; multiplies the value by one million</li>
     * <li>'t', 'T', 'k' or 'K' &mdash; multiplies the value by one thousand</li>
     * </ul>
     *
     * @param buffer    The text to process.
     * @param def       The default value to return, if the buffer cannot be parsed.
     * @param localized {@code true} if the text was localized.
     * @return The value.
     */
    public static final long extractLong(String buffer, long def, boolean localized) {
        buffer = normalizeNumber(buffer, localized);
        if (hasDecimalSeparator(buffer, localized)) {
            return (int) extractDouble(buffer, def, localized);
        }
        long multiplier = 1;
        if (hasBillionsSuffix(buffer)) {
            multiplier = 1000000000;
            buffer = removeSuffix(buffer);
        } else if (hasMillionsSuffix(buffer)) {
            multiplier = 1000000;
            buffer = removeSuffix(buffer);
        } else if (hasThousandsSuffix(buffer)) {
            multiplier = 1000;
            buffer = removeSuffix(buffer);
        }
        long max = Long.MAX_VALUE / multiplier;
        long min = Long.MIN_VALUE / multiplier;
        try {
            long value = Long.parseLong(buffer);
            if (value > max) {
                value = max;
            } else if (value < min) {
                value = min;
            }
            return value * multiplier;
        } catch (Exception exception) {
            return def;
        }
    }

    /**
     * Extracts a value from the specified buffer. In addition to typical input, this method can
     * also handle some suffixes:
     * <ul>
     * <li>'b', 'B', 'g' or 'G' &mdash; multiplies the value by one billion</li>
     * <li>'m' or 'M' &mdash; multiplies the value by one million</li>
     * <li>'t', 'T', 'k' or 'K' &mdash; multiplies the value by one thousand</li>
     * </ul>
     *
     * @param buffer    The text to process.
     * @param def       The default value to return, if the buffer cannot be parsed.
     * @param min       The minimum value to return.
     * @param max       The maximum value to return.
     * @param localized {@code true} if the text was localized.
     * @return The value.
     */
    public static final long extractLong(String buffer, long def, long min, long max, boolean localized) {
        return Math.min(Math.max(extractLong(buffer, def, localized), min), max);
    }

    /**
     * Extracts a value from the specified buffer. In addition to typical input, this method can
     * also handle some suffixes:
     * <ul>
     * <li>'b', 'B', 'g' or 'G' &mdash; multiplies the value by one billion</li>
     * <li>'m' or 'M' &mdash; multiplies the value by one million</li>
     * <li>'t', 'T', 'k' or 'K' &mdash; multiplies the value by one thousand</li>
     * </ul>
     *
     * @param buffer    The text to process.
     * @param def       The default value to return, if the buffer cannot be parsed.
     * @param localized {@code true} if the text was localized.
     * @return The value.
     */
    public static final double extractDouble(String buffer, double def, boolean localized) {
        buffer = normalizeNumber(buffer, localized);
        double multiplier = 1;
        if (hasBillionsSuffix(buffer)) {
            multiplier = 1000000000;
            buffer = removeSuffix(buffer);
        } else if (hasMillionsSuffix(buffer)) {
            multiplier = 1000000;
            buffer = removeSuffix(buffer);
        } else if (hasThousandsSuffix(buffer)) {
            multiplier = 1000;
            buffer = removeSuffix(buffer);
        }
        double max = Double.MAX_VALUE / multiplier;
        // NOTE: Do not use Double.MIN_VALUE here, as it isn't actually the minimum value... it is
        // merely the minimum POSITIVE value for some reason.
        double min = -max;
        try {
            if (localized) {
                char decimal = LOCALIZED_DECIMAL_SEPARATOR.charAt(0);
                if (decimal != '.') {
                    buffer = buffer.replace(decimal, '.');
                }
            }
            double value = Double.parseDouble(buffer);
            if (value > max) {
                value = max;
            } else if (value < min) {
                value = min;
            }
            return value * multiplier;
        } catch (Exception exception) {
            return def;
        }
    }

    /**
     * Extracts a value from the specified buffer. In addition to typical input, this method can
     * also handle some suffixes:
     * <ul>
     * <li>'b', 'B', 'g' or 'G' &mdash; multiplies the value by one billion</li>
     * <li>'m' or 'M' &mdash; multiplies the value by one million</li>
     * <li>'t', 'T', 'k' or 'K' &mdash; multiplies the value by one thousand</li>
     * </ul>
     *
     * @param buffer    The text to process.
     * @param def       The default value to return, if the buffer cannot be parsed.
     * @param min       The minimum value to return.
     * @param max       The maximum value to return.
     * @param localized {@code true} if the text was localized.
     * @return The value.
     */
    public static final double extractDouble(String buffer, double def, double min, double max, boolean localized) {
        return Math.min(Math.max(extractDouble(buffer, def, localized), min), max);
    }

    /**
     * @param buffer The string to convert.
     * @return The number of milliseconds since midnight, January 1, 1970.
     */
    public static long extractDate(String buffer) {
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
    public static long extractDateTime(String buffer) {
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
            return extractDate(buffer);
        }
        return System.currentTimeMillis();
    }

    /**
     * @param text      The text to process.
     * @param localized {@code true} if the text was localized.
     * @return The input, minus any trailing '0' characters. If at least one '0' was removed and the
     *         result would end with a '.' (or the localized equivalent, if {@code localized} is
     *         {@code true}), then the '.' is removed as well.
     */
    public static final String trimTrailingZeroes(String text, boolean localized) {
        if (text == null) {
            return null;
        }
        int dot = text.indexOf(localized ? LOCALIZED_DECIMAL_SEPARATOR.charAt(0) : '.');
        if (dot == -1) {
            return text;
        }
        int pos = text.length() - 1;
        if (dot == pos) {
            return text;
        }
        while (pos > dot && text.charAt(pos) == '0') {
            pos--;
        }
        if (dot == pos) {
            pos--;
        }
        return text.substring(0, pos + 1);
    }

    public static String normalizeNumber(String buffer, boolean localized) {
        if (buffer == null) {
            return "";
        }
        buffer = buffer.replaceAll(localized ? SAFE_LOCALIZED_GROUPING_SEPARATOR : ",", "").trim();
        if (!buffer.isEmpty() && buffer.charAt(0) == '+') {
            return buffer.substring(1).trim();
        }
        return buffer;
    }

    private static boolean hasDecimalSeparator(String buffer, boolean localized) {
        return buffer.indexOf(localized ? LOCALIZED_DECIMAL_SEPARATOR.charAt(0) : '.') != -1;
    }

    private static boolean hasBillionsSuffix(String buffer) {
        return buffer.endsWith("b") || buffer.endsWith("B") || buffer.endsWith("g") || buffer.endsWith("G");
    }

    private static boolean hasMillionsSuffix(String buffer) {
        return buffer.endsWith("m") || buffer.endsWith("M");
    }

    private static boolean hasThousandsSuffix(String buffer) {
        return buffer.endsWith("t") || buffer.endsWith("T") || buffer.endsWith("k") || buffer.endsWith("K");
    }

    private static String removeSuffix(String buffer) {
        return buffer.substring(0, buffer.length() - 1).trim();
    }

    /**
     * @param value The value to format.
     * @return The formatted value.
     */
    public static final String format(boolean value) {
        return value ? YES : NO;
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
    public static String formatWithForcedSign(long value) {
        return NUMBER_PLUS_FORMAT.format(value);
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
    public static String formatWithForcedSign(double value) {
        return NUMBER_PLUS_FORMAT.format(value);
    }
}
