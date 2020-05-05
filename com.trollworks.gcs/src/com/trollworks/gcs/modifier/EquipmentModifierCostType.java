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

/** Describes how an {@link EquipmentModifier}'s cost is applied. */
public enum EquipmentModifierCostType {
    /**
     * Modifies the original value stored in the equipment. Can be a ±value or a ±% value. Examples:
     * '+5', '-5', '+10%', '-10%'
     */
    TO_ORIGINAL_COST(ModifierValueType.ADDITION, ModifierValueType.PERCENTAGE) {
        @Override
        public String toShortString() {
            return "to original cost";
        }

        @Override
        public String toString() {
            return toShortString() + " (e.g. '+5', '-5', '+10%', '-10%')";
        }
    },
    /**
     * Modifies the base cost. Can be an additive multiplier or a CF value. Examples: 'x2', '+2 CF',
     * '-2 CF'
     */
    TO_BASE_COST(ModifierValueType.CF, ModifierValueType.MULTIPLIER) {
        @Override
        public String toShortString() {
            return "to base cost";
        }

        @Override
        public String toString() {
            return toShortString() + " (e.g. 'x2', '+2 CF', '-2 CF')";
        }
    },
    /**
     * Modifies the final base cost. Can be a ±value or a ±% value. Examples: '+5', '-5', '+10%',
     * '-10%'
     */
    TO_FINAL_BASE_COST(ModifierValueType.ADDITION, ModifierValueType.PERCENTAGE) {
        @Override
        public String toShortString() {
            return "to final base cost";
        }

        @Override
        public String toString() {
            return toShortString() + " (e.g. '+5', '-5', '+10%', '-10%')";
        }
    },
    /**
     * Modifies the final cost. Can be a ±value or a ±% value. Examples: '+5', '-5', '+10%', '-10%'
     */
    TO_FINAL_COST(ModifierValueType.ADDITION, ModifierValueType.PERCENTAGE) {
        @Override
        public String toShortString() {
            return "to final cost";
        }

        @Override
        public String toString() {
            return toShortString() + " (e.g. '+5', '-5', '+10%', '-10%')";
        }
    };

    private ModifierValueType[] mPermittedTypes;

    EquipmentModifierCostType(ModifierValueType... permittedTypes) {
        mPermittedTypes = permittedTypes;
    }

    public abstract String toShortString();

    public ModifierValueType determineType(String text) {
        ModifierValueType mvt = ModifierValueType.determineType(text);
        return allowed(mvt) ? mvt : mPermittedTypes[0];
    }

    public Fixed6 extractValue(String text, boolean localized) {
        return determineType(text).extractValue(text, localized);
    }

    public String format(String text, boolean localized) {
        ModifierValueType mvt = determineType(text);
        return mvt.format(mvt.extractValue(text, localized), localized);
    }

    private boolean allowed(ModifierValueType mvt) {
        for (ModifierValueType allowed : mPermittedTypes) {
            if (allowed == mvt) {
                return true;
            }
        }
        return false;
    }
}
