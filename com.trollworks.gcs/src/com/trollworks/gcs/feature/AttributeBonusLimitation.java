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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.I18n;

/** The limitations applicable to a {@link AttributeBonus}. */
public enum AttributeBonusLimitation {
    /** No limitation. */
    NONE {
        @Override
        public String toString() {
            return " ";
        }
    },
    /** Striking only. */
    STRIKING_ONLY {
        @Override
        public String toString() {
            return I18n.Text("for striking only");
        }
    },
    /** Lifting only */
    LIFTING_ONLY {
        @Override
        public String toString() {
            return I18n.Text("for lifting only");
        }
    }
}
