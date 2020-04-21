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

import java.text.ParseException;
import javax.swing.JFormattedTextField;

/** Provides integer field conversion. */
public class DoubleFormatter extends JFormattedTextField.AbstractFormatter {
    private double  mMinValue;
    private double  mMaxValue;
    private boolean mForceSign;

    /**
     * Creates a new {@link DoubleFormatter}.
     *
     * @param minValue  The minimum value allowed.
     * @param maxValue  The maximum value allowed.
     * @param forceSign Whether or not a plus sign should be forced for positive numbers.
     */
    public DoubleFormatter(double minValue, double maxValue, boolean forceSign) {
        mMinValue = minValue;
        mMaxValue = maxValue;
        mForceSign = forceSign;
    }

    @Override
    public Object stringToValue(String text) throws ParseException {
        return Double.valueOf(Math.min(Math.max(Numbers.extractDouble(text, mMinValue <= 0 && mMaxValue >= 0 ? 0 : mMinValue, true), mMinValue), mMaxValue));
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        double val = ((Double) value).doubleValue();
        return mForceSign ? Numbers.formatWithForcedSign(val) : Numbers.format(val);
    }
}
