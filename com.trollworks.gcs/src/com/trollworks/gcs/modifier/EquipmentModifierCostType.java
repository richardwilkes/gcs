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

import com.trollworks.gcs.utility.Fixed4;
import com.trollworks.gcs.utility.I18n;

/**
 * Describes how an {@link EquipmentModifier}'s cost is applied. These should be applied from top to
 * bottom, with any values of the same type adding together before being applied.
 */
public enum EquipmentModifierCostType {
    /** Adds to the base cost. */
    BASE_ADDITION {
        @Override
        public String toString() {
            return I18n.Text("to base cost");
        }

        @Override
        public String format(Fixed4 value) {
            String str = value.toLocalizedString();
            return value.greaterThanOrEqual(Fixed4.ZERO) ? "+" + str : str;
        }

        @Override
        public Fixed4 extract(String text, boolean localized) {
            return new Fixed4(text, Fixed4.ZERO, localized);
        }

        @Override
        public String adjustText(String text) {
            return ensureTextStartsWithPlusOrMinus(text);
        }
    },
    /** Multiplies the base cost. */
    BASE_MULTIPLIER {
        @Override
        public String toString() {
            return "× base cost";
        }

        @Override
        public String format(Fixed4 value) {
            return value.toLocalizedString();
        }

        @Override
        public Fixed4 extract(String text, boolean localized) {
            Fixed4 value = new Fixed4(text, Fixed4.ZERO, localized);
            return value.lessThan(Fixed4.ZERO) ? Fixed4.ONE : value;
        }

        @Override
        public String adjustText(String text) {
            return adjustTextForMultiplier(text);
        }
    },
    /** Adds to the cost factor. */
    COST_FACTOR {
        @Override
        public String toString() {
            return "CF";
        }

        @Override
        public String format(Fixed4 value) {
            String str = value.toLocalizedString();
            return value.greaterThanOrEqual(Fixed4.ZERO) ? "+" + str : str;
        }

        @Override
        public Fixed4 extract(String text, boolean localized) {
            return new Fixed4(text, Fixed4.ZERO, localized);
        }

        @Override
        public String adjustText(String text) {
            return ensureTextStartsWithPlusOrMinus(text);
        }
    },
    /** Multiplies the final cost. */
    FINAL_MULTIPLIER {
        @Override
        public String toString() {
            return "× final cost";
        }

        @Override
        public String format(Fixed4 value) {
            return value.toLocalizedString();
        }

        @Override
        public Fixed4 extract(String text, boolean localized) {
            Fixed4 value = new Fixed4(text, Fixed4.ZERO, localized);
            return value.lessThan(Fixed4.ZERO) ? Fixed4.ONE : value;
        }

        @Override
        public String adjustText(String text) {
            return adjustTextForMultiplier(text);
        }
    },
    /** Adds to the final cost. */
    FINAL_ADDITION {
        @Override
        public String toString() {
            return "to final cost";
        }

        @Override
        public String format(Fixed4 value) {
            String str = value.toLocalizedString();
            return value.greaterThanOrEqual(Fixed4.ZERO) ? "+" + str : str;
        }

        @Override
        public Fixed4 extract(String text, boolean localized) {
            return new Fixed4(text, Fixed4.ZERO, localized);
        }

        @Override
        public String adjustText(String text) {
            return ensureTextStartsWithPlusOrMinus(text);
        }
    };

    public abstract String format(Fixed4 value);

    public abstract Fixed4 extract(String text, boolean localized);

    public abstract String adjustText(String text);

    String ensureTextStartsWithPlusOrMinus(String text) {
        if (!text.startsWith("+") && !text.startsWith("-")) {
            return "+" + text;
        }
        return text;
    }

    String adjustTextForMultiplier(String text) {
        Fixed4 value = new Fixed4(text, Fixed4.ZERO, true);
        if (value.lessThanOrEqual(Fixed4.ZERO)) {
            return "1";
        }
        if (text.startsWith("+")) {
            return text.substring(1);
        }
        return text;
    }
}
