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

import java.awt.print.PageFormat;
import javax.print.attribute.standard.OrientationRequested;

/** Constants representing the various page orientation possibilities. */
public enum PageOrientation {
    /** Maps to {@link OrientationRequested#PORTRAIT}. */
    PORTRAIT {
        @Override
        public OrientationRequested getOrientationRequested() {
            return OrientationRequested.PORTRAIT;
        }

        @Override
        public String toString() {
            return I18n.Text("Portrait");
        }
    },
    /** Maps to {@link OrientationRequested#LANDSCAPE}. */
    LANDSCAPE {
        @Override
        public OrientationRequested getOrientationRequested() {
            return OrientationRequested.LANDSCAPE;
        }

        @Override
        public String toString() {
            return I18n.Text("Landscape");
        }
    },
    /** Maps to {@link OrientationRequested#REVERSE_PORTRAIT}. */
    REVERSE_PORTRAIT {
        @Override
        public OrientationRequested getOrientationRequested() {
            return OrientationRequested.REVERSE_PORTRAIT;
        }

        @Override
        public String toString() {
            return I18n.Text("Reversed Portrait");
        }
    },
    /** Maps to {@link OrientationRequested#REVERSE_LANDSCAPE}. */
    REVERSE_LANDSCAPE {
        @Override
        public OrientationRequested getOrientationRequested() {
            return OrientationRequested.REVERSE_LANDSCAPE;
        }

        @Override
        public String toString() {
            return I18n.Text("Reversed Landscape");
        }
    };

    /** @return The orientation attribute. */
    public abstract OrientationRequested getOrientationRequested();

    /**
     * @param orientation The {@link OrientationRequested} to load from.
     * @return The page orientation.
     */
    public static final PageOrientation get(OrientationRequested orientation) {
        for (PageOrientation one : values()) {
            if (one.getOrientationRequested() == orientation) {
                return one;
            }
        }
        return PORTRAIT;
    }

    /**
     * @param format The {@link PageFormat} to load from.
     * @return The page orientation.
     */
    public static final PageOrientation get(PageFormat format) {
        switch (format.getOrientation()) {
        case PageFormat.LANDSCAPE:
            return LANDSCAPE;
        case PageFormat.REVERSE_LANDSCAPE:
            return REVERSE_LANDSCAPE;
        case PageFormat.PORTRAIT:
        default:
            return PORTRAIT;
        }
    }
}
