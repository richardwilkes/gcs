/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.units.WeightValue;

import java.text.ParseException;

import javax.swing.JFormattedTextField;

/** Provides weight reduction field conversion. */
public class WeightReductionFormatter extends JFormattedTextField.AbstractFormatter {
    @Override
    public Object stringToValue(String text) throws ParseException {
        text = text != null ? text.trim() : ""; //$NON-NLS-1$
        if (text.endsWith("%")) { //$NON-NLS-1$
            return Integer.valueOf(Numbers.extractInteger(text.substring(0, text.length() - 1), 0, true));
        }
        return WeightValue.extract(text, true);
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        if (value instanceof Integer) {
            int percentage = ((Integer) value).intValue();
            if (percentage != 0) {
                return Numbers.format(percentage) + "%"; //$NON-NLS-1$
            }
            return ""; //$NON-NLS-1$
        } else if (value instanceof WeightValue) {
            WeightValue weight = (WeightValue) value;
            if (weight.getValue() == 0) {
                return ""; //$NON-NLS-1$
            }
            return weight.toString();
        }
        return ""; //$NON-NLS-1$
    }
}
