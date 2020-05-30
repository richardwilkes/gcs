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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.WeightUnits;

public enum ModifierWeightValueType {
    ADDITION {
        @Override
        public String format(Fraction fraction, boolean localized) {
            String str = localized ? fraction.toLocalizedString() : fraction.toString();
            return fraction.mNumerator.greaterThanOrEqual(Fixed6.ZERO) ? "+" + str : str;
        }
    }, PERCENTAGE_ADDER {
        @Override
        public String format(Fraction fraction, boolean localized) {
            String str = (localized ? fraction.toLocalizedString() : fraction.toString()) + "%";
            return fraction.mNumerator.greaterThanOrEqual(Fixed6.ZERO) ? "+" + str : str;
        }
    }, PERCENTAGE_MULTIPLIER {
        @Override
        Fraction adjustFraction(Fraction fraction) {
            return fraction.mNumerator.lessThanOrEqual(Fixed6.ZERO) ? new Fraction(new Fixed6(100), Fixed6.ONE) : fraction;
        }

        @Override
        public String format(Fraction fraction, boolean localized) {
            fraction = adjustFraction(fraction);
            return "x" + (localized ? fraction.toLocalizedString() : fraction.toString()) + "%";
        }
    }, MULTIPLIER {
        @Override
        Fraction adjustFraction(Fraction fraction) {
            return fraction.mNumerator.lessThanOrEqual(Fixed6.ZERO) ? new Fraction(Fixed6.ONE, Fixed6.ONE) : fraction;
        }

        @Override
        public String format(Fraction fraction, boolean localized) {
            fraction = adjustFraction(fraction);
            return "x" + (localized ? fraction.toLocalizedString() : fraction.toString());
        }
    };

    Fraction adjustFraction(Fraction fraction) {
        return fraction;
    }

    public abstract String format(Fraction fraction, boolean localized);

    public Fraction extractFraction(String in, boolean localized) {
        return adjustFraction(new Fraction(in, localized));
    }

    public static WeightUnits extractUnits(String in, WeightUnits def) {
        in = in.trim();
        for (WeightUnits units : WeightUnits.values()) {
            String text = Enums.toId(units);
            if (in.endsWith(text)) {
                return units;
            }
        }
        return def;
    }

    public static ModifierWeightValueType determineType(String in) {
        in = in.trim().toLowerCase();
        if (in.endsWith("%")) {
            if (in.startsWith("x")) {
                return PERCENTAGE_MULTIPLIER;
            }
            return PERCENTAGE_ADDER;
        }
        if (in.startsWith("x") || in.endsWith("x")) {
            return MULTIPLIER;
        }
        return ADDITION;
    }
}
