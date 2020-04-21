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

public enum Levels {
    NO_LEVELS {
        @Override
        public String toString() {
            return I18n.Text("Has No Levels");
        }
    }, HAS_LEVELS {
        @Override
        public String toString() {
            return I18n.Text("Has Levels");
        }
    }, HAS_HALF_LEVELS {
        @Override
        public String toString() {
            return I18n.Text("Has Half Levels");
        }
    }
}
