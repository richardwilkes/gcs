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
import java.text.ParseException;
import java.util.Date;
import javax.swing.JFormattedTextField;

/** Provides date field conversion. */
public class DateFormatter extends JFormattedTextField.AbstractFormatter {
    private int mType;

    /**
     * Creates a new {@link DateFormatter}.
     *
     * @param type The type of date format to use, one of {@link DateFormat#SHORT}, {@link
     *             DateFormat#MEDIUM}, {@link DateFormat#LONG}, or {@link DateFormat#FULL}.
     */
    public DateFormatter(int type) {
        mType = type;
    }

    @Override
    public Object stringToValue(String text) throws ParseException {
        return Long.valueOf(Numbers.extractDate(text));
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        Date date = new Date(((Long) value).longValue());
        return DateFormat.getDateInstance(mType).format(date);
    }
}
