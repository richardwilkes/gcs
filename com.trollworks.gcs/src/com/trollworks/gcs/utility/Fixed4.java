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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.utility.text.Numbers;

// Fixed4 holds a fixed-point value that contains up to 4 decimal places. Values are truncated, not
// rounded.
public class Fixed4 implements Comparable<Fixed4> {
    public static final Fixed4 ZERO = new Fixed4(0);
    public static final Fixed4 ONE  = new Fixed4(1);
    private             long   mRawValue;

    private Fixed4(long value, boolean unused) {
        mRawValue = value;
    }

    public Fixed4(double value) {
        mRawValue = (long) (value * 10000);
    }

    public Fixed4(long value) {
        mRawValue = value * 10000;
    }

    public Fixed4(String in, boolean localized) throws NumberFormatException {
        if (in == null || in.isBlank()) {
            throw new NumberFormatException("empty or null string is not valid");
        }
        in = Numbers.normalizeNumber(in, localized);
        if (localized) {
            char decimal = Numbers.LOCALIZED_DECIMAL_SEPARATOR.charAt(0);
            if (decimal != '.') {
                in = in.replace(decimal, '.');
            }
        }
        if (in.contains("E") || in.contains("e")) {
            // Given a floating-point value with an exponent, which technically isn't valid input,
            // but we'll try to convert it anyway.
            mRawValue = (long) (Double.parseDouble(in) * 10000);
        } else {
            boolean  neg   = false;
            String[] parts = in.split("\\.", 2);
            if (!parts[0].isBlank()) {
                if ("-".equals(parts[0]) || "-0".equals(parts[0])) {
                    neg = true;
                } else {
                    mRawValue = Long.parseLong(parts[0]) * 10000;
                    if (mRawValue < 0) {
                        neg = true;
                        mRawValue = -mRawValue;
                    }
                }
            }
            if (parts.length > 1) {
                StringBuilder buffer = new StringBuilder();
                buffer.append('1');
                buffer.append(parts[1]);
                if (buffer.length() > 5) {
                    buffer.setLength(5);
                } else {
                    while (buffer.length() < 5) {
                        buffer.append('0');
                    }
                }
                mRawValue += Long.parseLong(buffer.toString()) - 10000;
            }
            if (neg) {
                mRawValue = -mRawValue;
            }
        }
    }

    public Fixed4(String in, Fixed4 def, boolean localized) {
        try {
            mRawValue = new Fixed4(in, localized).mRawValue;
        } catch (Exception exception) {
            mRawValue = def.mRawValue;
        }
    }

    public Fixed4 add(Fixed4 other) {
        return new Fixed4(mRawValue + other.mRawValue, true);
    }

    public Fixed4 sub(Fixed4 other) {
        return new Fixed4(mRawValue - other.mRawValue, true);
    }

    public Fixed4 mul(Fixed4 other) {
        return new Fixed4((mRawValue * other.mRawValue) / 10000, true);
    }

    public Fixed4 div(Fixed4 other) {
        return new Fixed4((mRawValue * 10000) / other.mRawValue, true);
    }

    /** @return a new value which has everything to the right of the decimal place truncated */
    public Fixed4 trunc() {
        return new Fixed4((mRawValue / 10000) * 10000, true);
    }

    public long asLong() {
        return mRawValue / 10000;
    }

    public double asDouble() {
        return ((double) mRawValue) / 10000;
    }

    public boolean lessThan(Fixed4 other) {
        return mRawValue < other.mRawValue;
    }

    public boolean lessThanOrEqual(Fixed4 other) {
        return mRawValue <= other.mRawValue;
    }

    public boolean greaterThan(Fixed4 other) {
        return mRawValue > other.mRawValue;
    }

    public boolean greaterThanOrEqual(Fixed4 other) {
        return mRawValue >= other.mRawValue;
    }

    @Override
    public int compareTo(Fixed4 other) {
        return Long.compare(mRawValue, other.mRawValue);
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        return mRawValue == ((Fixed4) other).mRawValue;
    }

    @Override
    public int hashCode() {
        return Long.hashCode(mRawValue);
    }

    /** @return the same as toString(), but localized */
    public String toLocalizedString() {
        long whole    = mRawValue / 10000;
        long fraction = mRawValue % 10000;
        if (fraction == 0) {
            return Numbers.format(whole);
        }
        if (fraction < 0) {
            fraction = -fraction;
        }
        fraction += 10000;
        String str = Long.toString(fraction);
        while (str.endsWith("0")) {
            str = str.substring(0, str.length() - 1);
        }
        StringBuilder buffer = new StringBuilder();
        if (whole == 0 && mRawValue < 0) {
            buffer.append('-');
        }
        buffer.append(Numbers.format(whole));
        buffer.append(Numbers.LOCALIZED_DECIMAL_SEPARATOR);
        buffer.append(str.substring(1));
        return buffer.toString();
    }

    public String toString() {
        long whole    = mRawValue / 10000;
        long fraction = mRawValue % 10000;
        if (fraction == 0) {
            return Long.toString(whole);
        }
        if (fraction < 0) {
            fraction = -fraction;
        }
        fraction += 10000;
        String str = Long.toString(fraction);
        while (str.endsWith("0")) {
            str = str.substring(0, str.length() - 1);
        }
        StringBuilder buffer = new StringBuilder();
        if (whole == 0 && mRawValue < 0) {
            buffer.append('-');
        }
        buffer.append(Long.toString(whole));
        buffer.append('.');
        buffer.append(str.substring(1));
        return buffer.toString();
    }
}
