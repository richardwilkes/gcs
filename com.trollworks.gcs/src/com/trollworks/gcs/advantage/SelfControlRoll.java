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
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLWriter;

/** The possible self-control rolls, from page B121. */
public enum SelfControlRoll {
    /** Never. */
    NOT_APPLICABLE {
        @Override
        public String toString() {
            return I18n.Text("CR: N/A (Cannot Resist)");
        }

        @Override
        public double getMultiplier() {
            return 2.5;
        }

        @Override
        public int getCR() {
            return 0;
        }
    },
    /** Rarely. */
    CR6 {
        @Override
        public String toString() {
            return I18n.Text("CR: 6 (Resist Rarely)");
        }

        @Override
        public double getMultiplier() {
            return 2;
        }

        @Override
        public int getCR() {
            return 6;
        }
    },
    /** Fairly often. */
    CR9 {
        @Override
        public String toString() {
            return I18n.Text("CR: 9 (Resist Fairly Often)");
        }

        @Override
        public double getMultiplier() {
            return 1.5;
        }

        @Override
        public int getCR() {
            return 9;
        }
    },
    /** Quite often. */
    CR12 {
        @Override
        public String toString() {
            return I18n.Text("CR: 12 (Resist Quite Often)");
        }

        @Override
        public double getMultiplier() {
            return 1;
        }

        @Override
        public int getCR() {
            return 12;
        }
    },
    /** Almost all the time. */
    CR15 {
        @Override
        public String toString() {
            return I18n.Text("CR: 15 (Resist Almost All The Time)");
        }

        @Override
        public double getMultiplier() {
            return 0.5;
        }

        @Override
        public int getCR() {
            return 15;
        }
    },
    /** No self-control roll. */
    NONE_REQUIRED {
        @Override
        public String toString() {
            return I18n.Text("None Required");
        }

        @Override
        public double getMultiplier() {
            return 1;
        }

        @Override
        public int getCR() {
            return Integer.MAX_VALUE;
        }

        @Override
        public String getDescriptionWithCost() {
            return "";
        }

        @Override
        public void save(XMLWriter out, String tag, SelfControlRollAdjustments adj) {
            // Do nothing.
        }
    };

    /** The attribute tag use for {@link SelfControlRollAdjustments}. */
    public static final String ATTR_ADJUSTMENT = "adj";

    /**
     * @param tagValue The value within a tag representing a SelfControlRoll.
     * @return The actual SelfControlRoll.
     */
    public static final SelfControlRoll get(String tagValue) {
        int value = Numbers.extractInteger(tagValue, Integer.MAX_VALUE, false);
        for (SelfControlRoll cr : values()) {
            if (cr.getCR() == value) {
                return cr;
            }
        }
        return NONE_REQUIRED;
    }

    /**
     * @param value The value to look for.
     * @return The actual SelfControlRoll.
     */
    public static final SelfControlRoll getByCRValue(int value) {
        for (SelfControlRoll cr : values()) {
            if (cr.getCR() == value) {
                return cr;
            }
        }
        return NONE_REQUIRED;
    }

    /** @return The description, along with the cost. */
    public String getDescriptionWithCost() {
        return toString() + ", x" + getMultiplier();
    }

    /** @return The cost multiplier. */
    public abstract double getMultiplier();

    /** @return The minimum number to roll to retain control. */
    public abstract int getCR();

    /**
     * @param out The {@link XMLWriter} to use.
     * @param tag The XML tag to use.
     * @param adj The {@link SelfControlRollAdjustments} being used.
     */
    public void save(XMLWriter out, String tag, SelfControlRollAdjustments adj) {
        if (adj == SelfControlRollAdjustments.NONE) {
            out.simpleTag(tag, getCR());
        } else {
            out.simpleTagWithAttribute(tag, getCR(), ATTR_ADJUSTMENT, Enums.toId(adj));
        }
    }
}
