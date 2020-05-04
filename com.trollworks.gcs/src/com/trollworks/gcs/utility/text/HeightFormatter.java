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

import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.units.LengthValue;

import java.text.ParseException;
import javax.swing.JFormattedTextField;

/** Provides height field conversion. */
public class HeightFormatter extends JFormattedTextField.AbstractFormatter {
    private boolean mBlankOnZero;

    /**
     * @param blankOnZero When {@code true}, a value of zero resolves to the empty string when
     *                    calling {@link #valueToString(Object)}.
     */
    public HeightFormatter(boolean blankOnZero) {
        mBlankOnZero = blankOnZero;
    }

    @Override
    public Object stringToValue(String text) throws ParseException {
        return LengthValue.extract(text, true);
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        LengthValue length = (LengthValue) value;
        if (mBlankOnZero && length.getValue().equals(Fixed6.ZERO)) {
            return "";
        }
        return length.toString();
    }
}
