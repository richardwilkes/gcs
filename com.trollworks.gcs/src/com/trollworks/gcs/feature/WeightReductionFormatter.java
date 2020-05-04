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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.WeightValue;

import java.text.ParseException;
import javax.swing.JFormattedTextField;

/** Provides weight reduction field conversion. */
public class WeightReductionFormatter extends JFormattedTextField.AbstractFormatter {
    @Override
    public Object stringToValue(String text) throws ParseException {
        text = text != null ? text.trim() : "";
        if (text.endsWith("%")) {
            return Integer.valueOf(Numbers.extractInteger(text.substring(0, text.length() - 1), 0, true));
        }
        return WeightValue.extract(text, true);
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        if (value instanceof Integer) {
            int percentage = ((Integer) value).intValue();
            if (percentage != 0) {
                return Numbers.format(percentage) + "%";
            }
            return "";
        } else if (value instanceof WeightValue) {
            WeightValue weight = (WeightValue) value;
            if (weight.getValue().equals(Fixed6.ZERO)) {
                return "";
            }
            return weight.toString();
        }
        return "";
    }
}
