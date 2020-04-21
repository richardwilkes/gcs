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

import javax.print.attribute.standard.PrintQuality;

/** Constants representing the various print quality possibilities. */
public enum Quality {
    /** Maps to {@link PrintQuality#HIGH}. */
    HIGH {
        @Override
        public PrintQuality getPrintQuality() {
            return PrintQuality.HIGH;
        }

        @Override
        public String toString() {
            return I18n.Text("High");
        }
    },
    /** Maps to {@link PrintQuality#NORMAL}. */
    NORMAL {
        @Override
        public PrintQuality getPrintQuality() {
            return PrintQuality.NORMAL;
        }

        @Override
        public String toString() {
            return I18n.Text("Normal");
        }
    },
    /** Maps to {@link PrintQuality#DRAFT}. */
    DRAFT {
        @Override
        public PrintQuality getPrintQuality() {
            return PrintQuality.DRAFT;
        }

        @Override
        public String toString() {
            return I18n.Text("Draft");
        }
    };

    /** @return The print quality attribute. */
    public abstract PrintQuality getPrintQuality();

    /**
     * @param sides The {@link PrintQuality} to load from.
     * @return The sides.
     */
    public static final Quality get(PrintQuality sides) {
        for (Quality one : values()) {
            if (one.getPrintQuality() == sides) {
                return one;
            }
        }
        return NORMAL;
    }
}
