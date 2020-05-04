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

import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.text.Enums;

/** Holds a value and {@link LengthUnits} pair. */
public class LengthValue extends UnitsValue<LengthUnits> {
    protected static final Fixed6 METRIC_CONVERSION_FACTOR = new Fixed6(2.54);

    /**
     * @param buffer    The buffer to extract a {@link LengthValue} from.
     * @param localized {@code true} if the string might have localized notation within it.
     * @return The result.
     */
    public static LengthValue extract(String buffer, boolean localized) {
        LengthUnits units = LengthUnits.IN;
        if (buffer != null) {
            buffer = buffer.trim();
            // Check for the special case of FEET_AND_INCHES first
            int feetMark   = buffer.indexOf('\'');
            int inchesMark = buffer.indexOf('"');
            if (feetMark != -1 || inchesMark != -1) {
                if (feetMark == -1) {
                    return new LengthValue(new Fixed6(buffer.substring(0, inchesMark), Fixed6.ZERO, localized), LengthUnits.FT_IN);
                }
                Fixed6 inches = new Fixed6(buffer.substring(inchesMark != -1 && feetMark > inchesMark ? inchesMark + 1 : 0, feetMark), Fixed6.ZERO, localized).mul(new Fixed6(12));
                if (inchesMark != -1) {
                    inches = inches.add(new Fixed6(buffer.substring(feetMark < inchesMark ? feetMark + 1 : 0, inchesMark), Fixed6.ZERO, localized));
                }
                return new LengthValue(inches, LengthUnits.FT_IN);
            }
            for (LengthUnits lu : LengthUnits.values()) {
                String text = Enums.toId(lu);
                if (buffer.endsWith(text)) {
                    units = lu;
                    buffer = buffer.substring(0, buffer.length() - text.length());
                    break;
                }
            }
        }
        return new LengthValue(new Fixed6(buffer, Fixed6.ZERO, localized), units);
    }

    /**
     * Creates a new {@link LengthValue}.
     *
     * @param value The value to use.
     * @param units The {@link Units} to use.
     */
    public LengthValue(Fixed6 value, LengthUnits units) {
        super(value, units);
    }

    /**
     * Creates a new {@link LengthValue} from an existing one.
     *
     * @param other The {@link LengthValue} to clone.
     */
    public LengthValue(LengthValue other) {
        super(other);
    }

    /**
     * Creates a new {@link LengthValue} from an existing one and converts it to the given {@link
     * LengthUnits}.
     *
     * @param other The {@link LengthValue} to convert.
     * @param units The {@link LengthUnits} to use.
     */
    public LengthValue(LengthValue other, LengthUnits units) {
        super(other, units);
    }

    @Override
    public LengthUnits getDefaultUnits() {
        return LengthUnits.FT_IN;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        return obj instanceof LengthValue && super.equals(obj);
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }
}
