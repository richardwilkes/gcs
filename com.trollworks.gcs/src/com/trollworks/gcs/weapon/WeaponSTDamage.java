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

package com.trollworks.gcs.weapon;

/** The type of strength dice to add to damage. */
public enum WeaponSTDamage {
    NONE {
        @Override
        public String toString() {
            return "";
        }
    }, THR {
        @Override
        public String toString() {
            return "thr";
        }
    }, THR_LEVELED {
        @Override
        public String toString() {
            return "thr (leveled)";
        }
    }, SW {
        @Override
        public String toString() {
            return "sw";
        }
    }, SW_LEVELED {
        @Override
        public String toString() {
            return "sw (leveled)";
        }
    }
}
