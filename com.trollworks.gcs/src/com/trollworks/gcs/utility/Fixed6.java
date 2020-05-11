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

import java.math.BigInteger;

// Fixed6 holds a fixed-point value that contains up to 6 decimal places. Values are truncated, not
// rounded.
public class Fixed6 implements Comparable<Fixed6> {
    public static final  Fixed6     ZERO       = new Fixed6(0);
    public static final  Fixed6     ONE        = new Fixed6(1);
    public static final  Fixed6     MIN        = new Fixed6(Long.MAX_VALUE, true);
    public static final  Fixed6     MAX        = new Fixed6(Long.MIN_VALUE, true);
    private static final long       FACTOR     = 1000000;
    public static final  BigInteger BIG_FACTOR = BigInteger.valueOf(FACTOR);
    private              long       mRawValue;

    private Fixed6(long value, boolean unused) {
        mRawValue = value;
    }

    public Fixed6(double value) {
        mRawValue = (long) (value * FACTOR);
    }

    public Fixed6(long value) {
        mRawValue = value * FACTOR;
    }

    public Fixed6(String in, boolean localized) throws NumberFormatException {
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
            mRawValue = (long) (Double.parseDouble(in) * FACTOR);
        } else {
            boolean  neg   = false;
            String[] parts = in.split("\\.", 2);
            if (!parts[0].isBlank()) {
                if ("-".equals(parts[0]) || "-0".equals(parts[0])) {
                    neg = true;
                } else {
                    mRawValue = Long.parseLong(parts[0]) * FACTOR;
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
                if (buffer.length() > 7) {
                    buffer.setLength(7);
                } else {
                    while (buffer.length() < 7) {
                        buffer.append('0');
                    }
                }
                mRawValue += Long.parseLong(buffer.toString()) - FACTOR;
            }
            if (neg) {
                mRawValue = -mRawValue;
            }
        }
    }

    public Fixed6(String in, Fixed6 def, boolean localized) {
        try {
            mRawValue = new Fixed6(in, localized).mRawValue;
        } catch (Exception exception) {
            mRawValue = def.mRawValue;
        }
    }

    public Fixed6 add(Fixed6 other) {
        return new Fixed6(mRawValue + other.mRawValue, true);
    }

    public Fixed6 sub(Fixed6 other) {
        return new Fixed6(mRawValue - other.mRawValue, true);
    }

    public Fixed6 mul(Fixed6 other) {
        // Use BigInteger here to allow cases that would normally overflow in the intermediate
        // stages to work
        return new Fixed6(BigInteger.valueOf(mRawValue).multiply(BigInteger.valueOf(other.mRawValue)).divide(BIG_FACTOR).longValue(), true);
    }

    public Fixed6 div(Fixed6 other) {
        // Use BigInteger here to allow cases that would normally overflow in the intermediate
        // stages to work
        return new Fixed6(BigInteger.valueOf(mRawValue).multiply(BIG_FACTOR).divide(BigInteger.valueOf(other.mRawValue)).longValue(), true);
    }

    /** @return a new value which has everything to the right of the decimal place truncated */
    public Fixed6 trunc() {
        return new Fixed6((mRawValue / FACTOR) * FACTOR, true);
    }

    public Fixed6 round() {
        long whole    = mRawValue / FACTOR;
        long fraction = mRawValue % FACTOR;
        if (fraction > ((FACTOR / 2) - 1)) {
            whole++;
        }
        return new Fixed6(whole * FACTOR, true);
    }

    public long asLong() {
        return mRawValue / FACTOR;
    }

    public double asDouble() {
        return ((double) mRawValue) / FACTOR;
    }

    public boolean lessThan(Fixed6 other) {
        return mRawValue < other.mRawValue;
    }

    public boolean lessThanOrEqual(Fixed6 other) {
        return mRawValue <= other.mRawValue;
    }

    public boolean greaterThan(Fixed6 other) {
        return mRawValue > other.mRawValue;
    }

    public boolean greaterThanOrEqual(Fixed6 other) {
        return mRawValue >= other.mRawValue;
    }

    @Override
    public int compareTo(Fixed6 other) {
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
        return mRawValue == ((Fixed6) other).mRawValue;
    }

    @Override
    public int hashCode() {
        return Long.hashCode(mRawValue);
    }

    /** @return the same as toString(), but localized */
    public String toLocalizedString() {
        long whole    = mRawValue / FACTOR;
        long fraction = mRawValue % FACTOR;
        if (fraction == 0) {
            return Numbers.format(whole);
        }
        if (fraction < 0) {
            fraction = -fraction;
        }
        fraction += FACTOR;
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
        long whole    = mRawValue / FACTOR;
        long fraction = mRawValue % FACTOR;
        if (fraction == 0) {
            return Long.toString(whole);
        }
        if (fraction < 0) {
            fraction = -fraction;
        }
        fraction += FACTOR;
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
