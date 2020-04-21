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

package com.trollworks.gcs.utility.units;

import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;

/** Holds a value and {@link WeightUnits} pair. */
public class WeightValue extends UnitsValue<WeightUnits> {
    /**
     * @param buffer    The buffer to extract a {@link WeightValue} from.
     * @param localized {@code true} if the string might have localized notation within it.
     * @return The result.
     */
    public static WeightValue extract(String buffer, boolean localized) {
        WeightUnits units = WeightUnits.LB;
        if (buffer != null) {
            buffer = buffer.trim();
            for (WeightUnits lu : WeightUnits.values()) {
                String text = Enums.toId(lu);
                if (buffer.endsWith(text)) {
                    units = lu;
                    buffer = buffer.substring(0, buffer.length() - text.length());
                    break;
                }
            }
        }
        return new WeightValue(localized ? Numbers.extractDouble(buffer, 0, true) : Numbers.extractDouble(buffer, 0, false), units);
    }

    /**
     * Creates a new {@link WeightValue}.
     *
     * @param value The value to use.
     * @param units The {@link WeightUnits} to use.
     */
    public WeightValue(double value, WeightUnits units) {
        super(value, units);
    }

    /**
     * Creates a new {@link WeightValue} from an existing one.
     *
     * @param other The {@link WeightValue} to clone.
     */
    public WeightValue(WeightValue other) {
        super(other);
    }

    /**
     * Creates a new {@link WeightValue} from an existing one and converts it to the given {@link
     * WeightUnits}.
     *
     * @param other The {@link WeightValue} to convert.
     * @param units The {@link WeightUnits} to use.
     */
    public WeightValue(WeightValue other, WeightUnits units) {
        super(other, units);
    }

    @Override
    public WeightUnits getDefaultUnits() {
        return WeightUnits.LB;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        return obj instanceof WeightValue && super.equals(obj);
    }
}
