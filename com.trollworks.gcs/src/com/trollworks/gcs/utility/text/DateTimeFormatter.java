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

import javax.swing.JFormattedTextField;

/** Provides date field conversion. */
public class DateTimeFormatter extends JFormattedTextField.AbstractFormatter {
    @Override
    public Object stringToValue(String text) {
        return Long.valueOf(Numbers.extractDateTime(Numbers.DATE_AT_TIME_FORMAT, text));
    }

    @Override
    public String valueToString(Object value) {
        return Numbers.formatDateTime(Numbers.DATE_AT_TIME_FORMAT, ((Long) value).longValue());
    }
}
