/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.modifier;

import com.trollworks.toolkit.utility.I18n;

/**
 * Describes how a {@link EquipmentModifier}'s weight is applied. These should be applied from top
 * to bottom. Unlike with costs, all multipliers should be multiplied rather than added together
 * before being applied.
 */
public enum EquipmentModifierWeightType {
    /** Adds to the base weight. */
    BASE_ADDITION {
        @Override
        public String toString() {
            return I18n.Text("to base weight");
        }
    },
    /** Multiplies the weight. */
    MULTIPLIER {
        @Override
        public String toString() {
            return "\u00d7 weight";
        }
    },
    /** Adds to the final weight. */
    FINAL_ADDITION {
        @Override
        public String toString() {
            return I18n.Text("to final weight");
        }
    }
}
