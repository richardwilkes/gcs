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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.utility.I18n;

/** The types of {@link Advantage} containers. */
public enum AdvantageContainerType {
    /** The standard grouping container type. */
    GROUP {
        @Override
        public String toString() {
            return I18n.Text("Group");
        }
    },
    /**
     * The meta-trait grouping container type. Acts as one normal trait, listed as an advantage if
     * its point total is positive, or a disadvantage if it is negative.
     */
    META_TRAIT {
        @Override
        public String toString() {
            return I18n.Text("Meta-Trait");
        }
    },
    /**
     * The race grouping container type. Its point cost is tracked separately from normal advantages
     * and disadvantages.
     */
    RACE {
        @Override
        public String toString() {
            return I18n.Text("Race");
        }
    },
    /**
     * The alternative abilities grouping container type. It behaves similar to a {@link
     * #META_TRAIT} , but applies the rules for alternative abilities (see B61 and P11) to its
     * immediate children.
     */
    ALTERNATIVE_ABILITIES {
        @Override
        public String toString() {
            return I18n.Text("Alternative Abilities");
        }
    }
}
