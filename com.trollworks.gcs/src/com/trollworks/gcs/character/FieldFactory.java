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

import com.trollworks.gcs.utility.text.DateTimeFormatter;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.HeightFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.WeightFormatter;

import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

public final class FieldFactory {
    public static final long                    TIMESTAMP_FACTOR  = 60000; // milliseconds -> minutes
    public static final DefaultFormatterFactory DATETIME          = new DefaultFormatterFactory(new DateTimeFormatter(TIMESTAMP_FACTOR));
    public static final DefaultFormatterFactory HEIGHT            = new DefaultFormatterFactory(new HeightFormatter(true));
    public static final DefaultFormatterFactory WEIGHT            = new DefaultFormatterFactory(new WeightFormatter(true));
    public static final DefaultFormatterFactory SM                = new DefaultFormatterFactory(new IntegerFormatter(-99, 9999, true));
    public static final DefaultFormatterFactory PERCENT_REDUCTION = new DefaultFormatterFactory(new IntegerFormatter(0, 80, false));
    public static final DefaultFormatterFactory POSINT5           = new DefaultFormatterFactory(new IntegerFormatter(0, 99999, false));
    public static final DefaultFormatterFactory POSINT6           = new DefaultFormatterFactory(new IntegerFormatter(0, 999999, false));
    public static final DefaultFormatterFactory INT6              = new DefaultFormatterFactory(new IntegerFormatter(-999999, 999999, false));
    public static final DefaultFormatterFactory INT7              = new DefaultFormatterFactory(new IntegerFormatter(-9999999, 9999999, false));
    public static final DefaultFormatterFactory FLOAT             = new DefaultFormatterFactory(new DoubleFormatter(0, 99999, false));
    public static final DefaultFormatterFactory STRING;

    static {
        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        STRING = new DefaultFormatterFactory(formatter);
    }

    private FieldFactory() {
    }
}
