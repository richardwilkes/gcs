/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.text.DateTimeFormatter;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.Fixed6Formatter;
import com.trollworks.gcs.utility.text.HeightFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.WeightFormatter;

import javax.swing.JFormattedTextField;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

public final class FieldFactory {
    public static final long                    TIMESTAMP_FACTOR       = 60000; // milliseconds -> minutes
    public static final DefaultFormatterFactory BYTE                   = new DefaultFormatterFactory(new IntegerFormatter(0, 255, false));
    public static final DefaultFormatterFactory DATETIME               = new DefaultFormatterFactory(new DateTimeFormatter(TIMESTAMP_FACTOR));
    public static final DefaultFormatterFactory FIXED6                 = new DefaultFormatterFactory(new Fixed6Formatter(Fixed6.ZERO, new Fixed6(9999999999L), false));
    public static final DefaultFormatterFactory FLOAT                  = new DefaultFormatterFactory(new DoubleFormatter(0, 99999, false));
    public static final DefaultFormatterFactory HEIGHT                 = new DefaultFormatterFactory(new HeightFormatter(true));
    public static final DefaultFormatterFactory INT5                   = new DefaultFormatterFactory(new IntegerFormatter(-99999, 99999, false));
    public static final DefaultFormatterFactory INT6                   = new DefaultFormatterFactory(new IntegerFormatter(-999999, 999999, false));
    public static final DefaultFormatterFactory INT7                   = new DefaultFormatterFactory(new IntegerFormatter(-9999999, 9999999, false));
    public static final DefaultFormatterFactory INT9                   = new DefaultFormatterFactory(new IntegerFormatter(-999999999, 999999999, false));
    public static final DefaultFormatterFactory LENGTH                 = new DefaultFormatterFactory(new HeightFormatter(false));
    public static final DefaultFormatterFactory OUTPUT_DPI             = new DefaultFormatterFactory(new IntegerFormatter(50, 300, false));
    public static final DefaultFormatterFactory TOOLTIP_TIMEOUT        = new DefaultFormatterFactory(new IntegerFormatter(1, 300, false));
    public static final DefaultFormatterFactory TOOLTIP_MILLIS_TIMEOUT = new DefaultFormatterFactory(new IntegerFormatter(0, 5000, false));
    public static final DefaultFormatterFactory PERCENT_REDUCTION      = new DefaultFormatterFactory(new IntegerFormatter(0, 80, false));
    public static final DefaultFormatterFactory POSINT3                = new DefaultFormatterFactory(new IntegerFormatter(0, 999, false));
    public static final DefaultFormatterFactory POSINT5                = new DefaultFormatterFactory(new IntegerFormatter(0, 99999, false));
    public static final DefaultFormatterFactory POSINT6                = new DefaultFormatterFactory(new IntegerFormatter(0, 999999, false));
    public static final DefaultFormatterFactory POSINT9                = new DefaultFormatterFactory(new IntegerFormatter(0, 999999999, false));
    public static final DefaultFormatterFactory SINT2                  = new DefaultFormatterFactory(new IntegerFormatter(-99, 99, true));
    public static final DefaultFormatterFactory SM                     = new DefaultFormatterFactory(new IntegerFormatter(-99, 9999, true));
    public static final DefaultFormatterFactory STRING;
    public static final DefaultFormatterFactory WEIGHT                 = new DefaultFormatterFactory(new WeightFormatter(true));

    static {
        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        STRING = new DefaultFormatterFactory(formatter);
    }

    private FieldFactory() {
    }

    public static Integer getMaxValue(DefaultFormatterFactory factory) {
        JFormattedTextField.AbstractFormatter formatter = factory.getDefaultFormatter();
        if (formatter instanceof IntegerFormatter) {
            return Integer.valueOf(((IntegerFormatter) formatter).getMaxValue());
        }
        throw new RuntimeException("invalid factory for FieldFactory.getMaxValue()");
    }
}
