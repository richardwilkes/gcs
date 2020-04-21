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

/**
 * Describes how a {@link EquipmentModifier}'s cost is applied. These should be applied from top to
 * bottom, with any values of the same type adding together before being applied.
 */
public enum EquipmentModifierCostType {
    /** Adds to the base cost. */
    BASE_ADDITION {
        @Override
        public String toString() {
            return I18n.Text("to base cost");
        }
    },
    /** Multiplies the cost. */
    MULTIPLIER {
        @Override
        public String toString() {
            return "\u00d7 cost";
        }
    },
    /** Adds to the cost factor. */
    COST_FACTOR {
        @Override
        public String toString() {
            return "CF";
        }
    },
    /** Adds to the final cost. */
    FINAL_ADDITION {
        @Override
        public String toString() {
            return "to final cost";
        }
    }
}
