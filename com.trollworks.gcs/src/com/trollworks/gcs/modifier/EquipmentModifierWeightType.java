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
import com.trollworks.gcs.utility.units.WeightUnits;

/** Describes how a {@link EquipmentModifier}'s weight is applied. */
public enum EquipmentModifierWeightType {
    /**
     * Modifies the original value stored in the equipment. Can be a ±value or a ±% value. Examples:
     * '+5 lb', '-5 lb', '+10%', '-10%'
     */
    TO_ORIGINAL_WEIGHT(ModifierWeightValueType.ADDITION, ModifierWeightValueType.PERCENTAGE_ADDER) {
        @Override
        public String toShortString() {
            return I18n.Text("to original weight");
        }

        @Override
        public String toString() {
            return toShortString() + I18n.Text(" (e.g. '+5 lb', '-5 lb', '+10%', '-10%')");
        }
    },
    /**
     * Modifies the base cost. Can be a ±value, a ±% value, or a multiplier. Examples: '+5 lb',
     * '-5 lb', 'x10%', 'x3', 'x2/3'
     */
    TO_BASE_WEIGHT(ModifierWeightValueType.ADDITION, ModifierWeightValueType.PERCENTAGE_MULTIPLIER, ModifierWeightValueType.MULTIPLIER) {
        @Override
        public String toShortString() {
            return I18n.Text("to base weight");
        }

        @Override
        public String toString() {
            return toShortString() + I18n.Text(" (e.g. '+5 lb', '-5 lb', 'x10%', 'x3', 'x2/3')");
        }
    },
    /**
     * Modifies the final base cost. Can be a ±value, a ±% value, or a multiplier. Examples:
     * '+5 lb', '-5 lb', 'x10%', 'x3', 'x2/3'
     */
    TO_FINAL_BASE_WEIGHT(ModifierWeightValueType.ADDITION, ModifierWeightValueType.PERCENTAGE_MULTIPLIER, ModifierWeightValueType.MULTIPLIER) {
        @Override
        public String toShortString() {
            return I18n.Text("to final base weight");
        }

        @Override
        public String toString() {
            return toShortString() + I18n.Text(" (e.g. '+5 lb', '-5 lb', 'x10%', 'x3', 'x2/3')");
        }
    },
    /**
     * Modifies the final cost. Can be a ±value, a ±% value, or a multiplier. Examples: '+5 lb',
     * '-5 lb', 'x10%', 'x3', 'x2/3'
     */
    TO_FINAL_WEIGHT(ModifierWeightValueType.ADDITION, ModifierWeightValueType.PERCENTAGE_MULTIPLIER, ModifierWeightValueType.MULTIPLIER) {
        @Override
        public String toShortString() {
            return I18n.Text("to final weight");
        }

        @Override
        public String toString() {
            return toShortString() + I18n.Text(" (e.g. '+5 lb', '-5 lb', 'x10%', 'x3', 'x2/3')");
        }
    };

    private final ModifierWeightValueType[] mPermittedTypes;

    EquipmentModifierWeightType(ModifierWeightValueType... permittedTypes) {
        mPermittedTypes = permittedTypes;
    }

    public abstract String toShortString();

    public ModifierWeightValueType determineType(String text) {
        ModifierWeightValueType mvt = ModifierWeightValueType.determineType(text);
        return allowed(mvt) ? mvt : mPermittedTypes[0];
    }

    public Fraction extractValue(String text, boolean localized) {
        return determineType(text).extractFraction(text, localized);
    }

    public String format(String text, WeightUnits defUnits, boolean localized) {
        ModifierWeightValueType mvt = determineType(text);
        String                  str = mvt.format(mvt.extractFraction(text, localized), localized);
        if (mvt == ModifierWeightValueType.ADDITION) {
            str += " " + ModifierWeightValueType.extractUnits(text, defUnits).getAbbreviation();
        }
        return str;
    }

    private boolean allowed(ModifierWeightValueType mvt) {
        for (ModifierWeightValueType allowed : mPermittedTypes) {
            if (allowed == mvt) {
                return true;
            }
        }
        return false;
    }
}
