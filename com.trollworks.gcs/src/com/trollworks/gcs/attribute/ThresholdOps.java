/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.utility.I18n;

public enum ThresholdOps {
    UNKNOWN {
        @Override
        public String title() {
            return toString();
        }

        @Override
        public String toString() {
            return I18n.text("Unknown");
        }
    },
    HALVE_MOVE {
        @Override
        public String title() {
            return I18n.text("Halve Move");
        }

        @Override
        public String toString() {
            return I18n.text("Halve Move (round up)");
        }
    },
    HALVE_DODGE {
        @Override
        public String title() {
            return I18n.text("Halve Dodge");
        }

        @Override
        public String toString() {
            return I18n.text("Halve Dodge (round up)");
        }
    },
    HALVE_ST {
        @Override
        public String title() {
            return I18n.text("Halve ST");
        }

        @Override
        public String toString() {
            return I18n.text("Halve ST (round up; does not affect HP and damage)");
        }
    };

    abstract String title();
}
