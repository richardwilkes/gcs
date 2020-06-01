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

public enum WeaponSelectionType {
    THIS_WEAPON {
        @Override
        public String toString() {
            return I18n.Text("to this weapon");
        }
    }, WEAPONS_WITH_NAME {
        @Override
        public String toString() {
            return I18n.Text("to weapons whose name");
        }
    }, WEAPONS_WITH_REQUIRED_SKILL {
        @Override
        public String toString() {
            return I18n.Text("to weapons whose required skill name");
        }
    }
}
