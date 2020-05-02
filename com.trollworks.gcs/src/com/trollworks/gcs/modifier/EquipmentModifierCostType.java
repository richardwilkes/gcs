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

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

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
        public String format(double value) {
            return Numbers.formatWithForcedSign(value);
        }

        @Override
        public double extract(String text, boolean localized) {
            return Numbers.extractDouble(text, 0, localized);
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
        public String format(double value) {
            return Numbers.format(value);
        }

        @Override
        public double extract(String text, boolean localized) {
            double value = Numbers.extractDouble(text, 0, localized);
            return value < 0 ? 1 : value;
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
        public String format(double value) {
            return Numbers.formatWithForcedSign(value);
        }

        @Override
        public double extract(String text, boolean localized) {
            return Numbers.extractDouble(text, 0, localized);
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
        public String format(double value) {
            return Numbers.format(value);
        }

        @Override
        public double extract(String text, boolean localized) {
            double value = Numbers.extractDouble(text, 0, localized);
            return value < 0 ? 1 : value;
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
        public String format(double value) {
            return Numbers.formatWithForcedSign(value);
        }

        @Override
        public double extract(String text, boolean localized) {
            return Numbers.extractDouble(text, 0, localized);
        }

        @Override
        public String adjustText(String text) {
            return ensureTextStartsWithPlusOrMinus(text);
        }
    };

    public abstract String format(double value);

    public abstract double extract(String text, boolean localized);

    public abstract String adjustText(String text);

    String ensureTextStartsWithPlusOrMinus(String text) {
        if (!text.startsWith("+") && !text.startsWith("-")) {
            return "+" + text;
        }
        return text;
    }

    String adjustTextForMultiplier(String text) {
        double value = Numbers.extractDouble(text, 0, true);
        if (value <= 0) {
            return "1";
        }
        if (text.startsWith("+")) {
            return text.substring(1);
        }
        return text;
    }
}
