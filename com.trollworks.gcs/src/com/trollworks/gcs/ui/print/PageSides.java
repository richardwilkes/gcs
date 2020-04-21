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

import javax.print.attribute.standard.Sides;

/** Constants representing the various page side possibilities. */
public enum PageSides {
    /** Maps to {@link Sides#ONE_SIDED}. */
    SINGLE {
        @Override
        public Sides getSides() {
            return Sides.ONE_SIDED;
        }

        @Override
        public String toString() {
            return I18n.Text("Single");
        }
    },
    /** Maps to {@link Sides#DUPLEX}. */
    DUPLEX {
        @Override
        public Sides getSides() {
            return Sides.DUPLEX;
        }

        @Override
        public String toString() {
            return I18n.Text("Duplex");
        }
    },
    /** Maps to {@link Sides#TUMBLE}. */
    TUMBLE {
        @Override
        public Sides getSides() {
            return Sides.TUMBLE;
        }

        @Override
        public String toString() {
            return I18n.Text("Tumble");
        }
    };

    /** @return The sides attribute. */
    public abstract Sides getSides();

    /**
     * @param sides The {@link Sides} to load from.
     * @return The sides.
     */
    public static final PageSides get(Sides sides) {
        for (PageSides one : values()) {
            if (one.getSides() == sides) {
                return one;
            }
        }
        return SINGLE;
    }
}
