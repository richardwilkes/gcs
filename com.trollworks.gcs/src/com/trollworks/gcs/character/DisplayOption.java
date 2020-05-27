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

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.I18n;

public enum DisplayOption {
    NOT_SHOWN {
        @Override
        public String toString() {
            return I18n.Text("Not Shown");
        }
    }, INLINE {
        @Override
        public String toString() {
            return I18n.Text("Inline");
        }

        @Override
        public boolean inline() {
            return true;
        }
    }, TOOLTIP {
        @Override
        public String toString() {
            return I18n.Text("Tooltip");
        }

        @Override
        public boolean tooltip() {
            return true;
        }
    }, INLINE_AND_TOOLTIP {
        @Override
        public String toString() {
            return I18n.Text("Inline & Tooltip");
        }

        @Override
        public boolean inline() {
            return true;
        }

        @Override
        public boolean tooltip() {
            return true;
        }
    };

    public boolean inline() {
        return false;
    }

    public boolean tooltip() {
        return false;
    }
}
