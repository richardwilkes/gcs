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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.utility.Fixed6;

import java.text.DecimalFormatSymbols;

public class Fraction {
    public Fixed6 mNumerator;
    public Fixed6 mDenominator;

    public Fraction(String in, boolean localized) {
        String[] parts = in.split("/", 2);
        mNumerator = extractFixed6(parts[0], localized);
        if (parts.length > 1) {
            mDenominator = extractFixed6(parts[1], localized);
        } else {
            mDenominator = Fixed6.ONE;
        }
        normalize();
    }

    public Fraction(Fixed6 numerator, Fixed6 denominator) {
        mNumerator = numerator;
        mDenominator = denominator;
        normalize();
    }

    public void normalize() {
        if (mDenominator.equals(Fixed6.ZERO)) {
            mNumerator = Fixed6.ZERO;
            mDenominator = Fixed6.ONE;
        } else if (mDenominator.lessThan(Fixed6.ZERO)) {
            Fixed6 negOne = new Fixed6(-1);
            mNumerator = mNumerator.mul(negOne);
            mDenominator = mDenominator.mul(negOne);
        }
    }

    public Fixed6 value() {
        return mNumerator.div(mDenominator);
    }

    /** @return the same as toString(), but localized */
    public String toLocalizedString() {
        if (mDenominator.equals(Fixed6.ONE)) {
            return mNumerator.toLocalizedString();
        }
        return mNumerator.toLocalizedString() + "/" + mDenominator.toLocalizedString();
    }

    @Override
    public String toString() {
        if (mDenominator.equals(Fixed6.ONE)) {
            return mNumerator.toString();
        }
        return mNumerator.toString() + "/" + mDenominator.toString();
    }

    public static Fixed6 extractFixed6(String in, boolean localized) {
        in = in.trim();
        StringBuilder buffer        = new StringBuilder();
        char          decimal       = localized ? DecimalFormatSymbols.getInstance().getDecimalSeparator() : '.';
        char          group         = localized ? DecimalFormatSymbols.getInstance().getGroupingSeparator() : ',';
        int           length        = in.length();
        boolean       allowDecimal  = true;
        boolean       allowNegative = true;
        boolean       allowOther    = true;
        for (int i = 0; i < length; i++) {
            char ch = in.charAt(i);
            if (ch != group) {
                if (allowDecimal && ch == decimal) {
                    allowDecimal = false;
                    allowOther = false;
                    buffer.append('.');
                } else if (allowNegative && ch == '-') {
                    allowNegative = false;
                    allowOther = false;
                    buffer.append('-');
                } else if (ch >= '0' && ch <= '9') {
                    allowOther = false;
                    buffer.append(ch);
                } else if (!allowOther) {
                    break;
                }
            }
        }
        return new Fixed6(buffer.toString(), Fixed6.ZERO, false);
    }
}
