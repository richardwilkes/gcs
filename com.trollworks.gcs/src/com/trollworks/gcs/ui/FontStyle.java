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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.utility.I18n;

import java.awt.Font;

public enum FontStyle {
    PLAIN {
        @Override
        public String toString() {
            return I18n.text("Plain");
        }
    },
    BOLD {
        @Override
        public String toString() {
            return I18n.text("Bold");
        }
    },
    ITALIC {
        @Override
        public String toString() {
            return I18n.text("Italic");
        }
    },
    BOLD_ITALIC {
        @Override
        public String toString() {
            return I18n.text("Bold Italic");
        }
    };

    public static FontStyle from(Font font) {
        FontStyle[] styles = values();
        return styles[font.getStyle() % styles.length];
    }
}
