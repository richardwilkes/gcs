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

package com.trollworks.gcs.ui.print;

import com.trollworks.gcs.utility.I18n;

import javax.print.attribute.standard.Chromaticity;

/** Constants representing the various ink chromaticity possibilities. */
public enum InkChromaticity {
    /** Maps to {@link Chromaticity#COLOR}. */
    COLOR {
        @Override
        public Chromaticity getChromaticity() {
            return Chromaticity.COLOR;
        }

        @Override
        public String toString() {
            return I18n.Text("Color");
        }
    },
    /** Maps to {@link Chromaticity#MONOCHROME}. */
    MONOCHROME {
        @Override
        public Chromaticity getChromaticity() {
            return Chromaticity.MONOCHROME;
        }

        @Override
        public String toString() {
            return I18n.Text("Monochrome");
        }
    };

    /** @return The chromaticity attribute. */
    public abstract Chromaticity getChromaticity();

    /**
     * @param chromaticity The {@link Chromaticity} to load from.
     * @return The chromaticity.
     */
    public static final InkChromaticity get(Chromaticity chromaticity) {
        for (InkChromaticity one : values()) {
            if (one.getChromaticity() == chromaticity) {
                return one;
            }
        }
        return MONOCHROME;
    }
}
