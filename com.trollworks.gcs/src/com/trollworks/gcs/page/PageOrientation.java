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

package com.trollworks.gcs.page;

import com.trollworks.gcs.utility.I18n;

import java.awt.print.PageFormat;
import javax.print.attribute.standard.OrientationRequested;

/** Constants representing the page orientation possibilities. */
public enum PageOrientation {
    /** Maps to {@link OrientationRequested#PORTRAIT}. */
    PORTRAIT {
        public int getOrientationForPageFormat() {
            return PageFormat.PORTRAIT;
        }

        @Override
        public String toString() {
            return I18n.TextWithContext(1, "Portrait");
        }
    },
    /** Maps to {@link OrientationRequested#LANDSCAPE}. */
    LANDSCAPE {
        public int getOrientationForPageFormat() {
            return PageFormat.LANDSCAPE;
        }

        @Override
        public String toString() {
            return I18n.Text("Landscape");
        }
    };

    public abstract int getOrientationForPageFormat();
}
