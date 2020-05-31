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

import com.trollworks.gcs.utility.I18n;

import java.text.DateFormat;
import java.text.MessageFormat;
import java.text.ParseException;
import java.util.Date;
import javax.swing.JFormattedTextField;

/** Provides date field conversion. */
public class DateTimeFormatter extends JFormattedTextField.AbstractFormatter {
    @Override
    public Object stringToValue(String text) throws ParseException {
        return Long.valueOf(Numbers.extractDateTime(text));
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        return getFormattedDateTime(((Long) value).longValue());
    }

    public static String getFormattedDateTime(long dateTime) {
        Date date = new Date(dateTime);
        return MessageFormat.format(I18n.Text("{0} at {1}"), DateFormat.getDateInstance(DateFormat.MEDIUM).format(date), DateFormat.getTimeInstance(DateFormat.SHORT).format(date));
    }
}
