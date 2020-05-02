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

/** Describes how a {@link AdvantageModifier}'s point cost is applied. */
public enum AdvantageModifierCostType {
    /** Adds to the percentage multiplier. */
    PERCENTAGE {
        @Override
        public String toString() {
            return "%";
        }
    },
    /** Adds a constant to the base value prior to any multiplier or percentage adjustment. */
    POINTS {
        @Override
        public String toString() {
            return I18n.Text("points");
        }
    },
    /** Multiplies the final cost by a constant. */
    MULTIPLIER {
        @Override
        public String toString() {
            return "×";
        }
    }
}
