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

import java.text.DecimalFormatSymbols;

public enum ModifierValueType {
    ADDITION {
        @Override
        public String format(Fixed6 value, boolean localized) {
            String str = localized ? value.toLocalizedString() : value.toString();
            return value.greaterThanOrEqual(Fixed6.ZERO) ? "+" + str : str;
        }
    }, PERCENTAGE {
        @Override
        public String format(Fixed6 value, boolean localized) {
            String str = (localized ? value.toLocalizedString() : value.toString()) + "%";
            return value.greaterThanOrEqual(Fixed6.ZERO) ? "+" + str : str;
        }
    }, MULTIPLIER {
        @Override
        Fixed6 adjustValue(Fixed6 value) {
            return value.lessThanOrEqual(Fixed6.ZERO) ? Fixed6.ONE : value;
        }

        @Override
        public String format(Fixed6 value, boolean localized) {
            value = adjustValue(value);
            return "x" + (localized ? value.toLocalizedString() : value.toString());
        }
    }, CF {
        @Override
        public String format(Fixed6 value, boolean localized) {
            String str = (localized ? value.toLocalizedString() : value.toString()) + " CF";
            return value.greaterThanOrEqual(Fixed6.ZERO) ? "+" + str : str;
        }
    };

    Fixed6 adjustValue(Fixed6 value) {
        return value;
    }

    public abstract String format(Fixed6 value, boolean localized);

    public Fixed6 extractValue(String in, boolean localized) {
        in = in.trim();
        StringBuilder buffer        = new StringBuilder();
        char          decimal       = localized ? DecimalFormatSymbols.getInstance().getDecimalSeparator() : '.';
        char          group         = localized ? DecimalFormatSymbols.getInstance().getGroupingSeparator() : ',';
        int           length        = in.length();
        boolean       allowDecimal  = true;
        boolean       allowNegative = true;
        boolean       allowOther    = true;
        for (int i = 0; i < length; i++) {
            char ch = in.charAt(i);
            if (ch != group) {
                if (allowDecimal && ch == decimal) {
                    allowDecimal = false;
                    allowOther = false;
                    buffer.append('.');
                } else if (allowNegative && ch == '-') {
                    allowNegative = false;
                    allowOther = false;
                    buffer.append('-');
                } else if (ch >= '0' && ch <= '9') {
                    allowOther = false;
                    buffer.append(ch);
                } else if (!allowOther) {
                    break;
                }
            }
        }
        return adjustValue(new Fixed6(buffer.toString(), Fixed6.ZERO, false));
    }

    public static ModifierValueType determineType(String in) {
        in = in.trim().toLowerCase();
        if (in.endsWith("cf")) {
            return CF;
        }
        if (in.endsWith("%")) {
            return PERCENTAGE;
        }
        if (in.startsWith("x") || in.endsWith("x")) {
            return MULTIPLIER;
        }
        return ADDITION;
    }
}
