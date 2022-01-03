/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.text;

import com.trollworks.gcs.utility.Fixed6;

import javax.swing.JFormattedTextField;

/** Provides integer field conversion. */
public class Fixed6Formatter extends JFormattedTextField.AbstractFormatter {
    private Fixed6  mMinValue;
    private Fixed6  mMaxValue;
    private boolean mForceSign;

    /**
     * Creates a new Fixed6Formatter.
     *
     * @param minValue  The minimum value allowed.
     * @param maxValue  The maximum value allowed.
     * @param forceSign Whether or not a plus sign should be forced for positive numbers.
     */
    public Fixed6Formatter(Fixed6 minValue, Fixed6 maxValue, boolean forceSign) {
        mMinValue = minValue;
        mMaxValue = maxValue;
        mForceSign = forceSign;
    }

    @Override
    public Object stringToValue(String text) {
        Fixed6 value;
        try {
            value = new Fixed6(text, mMinValue.lessThanOrEqual(Fixed6.ZERO) && mMaxValue.greaterThanOrEqual(Fixed6.ZERO) ? Fixed6.ZERO : mMinValue, true);
        } catch (NumberFormatException nfe) {
            value = Fixed6.ZERO;
        }
        if (mMinValue.greaterThan(value)) {
            value = mMinValue;
        }
        if (mMaxValue.lessThan(value)) {
            value = mMaxValue;
        }
        return value;
    }

    @Override
    public String valueToString(Object value) {
        String result = ((Fixed6) value).toLocalizedString();
        if (mForceSign && !result.startsWith("-")) {
            result = "+" + result;
        }
        return result;
    }
}
