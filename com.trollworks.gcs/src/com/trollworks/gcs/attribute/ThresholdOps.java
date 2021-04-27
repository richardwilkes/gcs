/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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
            return I18n.Text("Unknown");
        }
    },
    HALVE_MOVE {
        @Override
        public String title() {
            return I18n.Text("Halve Move");
        }

        @Override
        public String toString() {
            return I18n.Text("Halve Move (round up)");
        }
    },
    HALVE_DODGE {
        @Override
        public String title() {
            return I18n.Text("Halve Dodge");
        }

        @Override
        public String toString() {
            return I18n.Text("Halve Dodge (round up)");
        }
    },
    HALVE_ST {
        @Override
        public String title() {
            return I18n.Text("Halve ST");
        }

        @Override
        public String toString() {
            return I18n.Text("Halve ST (round up; does not affect HP and damage)");
        }
    };

    abstract String title();
}
