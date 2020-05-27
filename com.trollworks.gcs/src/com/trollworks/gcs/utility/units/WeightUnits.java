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

package com.trollworks.gcs.utility.units;

import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;

import java.text.MessageFormat;

/** Common weight units. */
public enum WeightUnits implements Units {
    /** Ounces. */
    OZ(Fixed6.ONE.div(new Fixed6(16)), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Ounces");
        }
    },
    /** Pounds. */
    LB(Fixed6.ONE, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Pounds");
        }
    },
    /** Short Tons */
    TN(new Fixed6(2000), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Short Tons");
        }
    },
    /** Long Tons */
    LT(new Fixed6(2240), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Long Tons");
        }
    },
    /**
     * Metric Tons. Must come after Long Tons and Short Tons since it's abbreviation is a subset.
     */
    T(new Fixed6(2205), true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Metric Tons");
        }
    },
    /** Kilograms. */
    KG(new Fixed6("2.205", Fixed6.ZERO, false), true) { // entered as text to ensure precision
        @Override
        public String getLocalizedName() {
            return I18n.Text("Kilograms");
        }
    },
    /** Grams. Must come after Kilograms since it's abbreviation is a subset. */
    G(new Fixed6("0.002205", Fixed6.ZERO, false), true) { // entered as text to ensure precision
        @Override
        public String getLocalizedName() {
            return I18n.Text("Grams");
        }
    };

    private Fixed6  mFactor;
    private boolean mIsMetric;

    WeightUnits(Fixed6 factor, boolean isMetric) {
        mFactor = factor;
        mIsMetric = isMetric;
    }

    @Override
    public Fixed6 convert(Units units, Fixed6 value) {
        return units.getFactor().mul(value).div(mFactor);
    }

    @Override
    public Fixed6 normalize(Fixed6 value) {
        return mFactor.mul(value);
    }

    @Override
    public Fixed6 getFactor() {
        return mFactor;
    }

    @Override
    public String format(Fixed6 value, boolean localize) {
        return MessageFormat.format("{0} {1}", localize ? value.toLocalizedString() : value.toString(), getAbbreviation());
    }

    @Override
    public WeightUnits[] getCompatibleUnits() {
        return values();
    }

    @Override
    public String getAbbreviation() {
        return name().toLowerCase();
    }

    @Override
    public String toString() {
        return String.format("%s (%s)", getLocalizedName(), getAbbreviation());
    }

    public boolean isMetric() {
        return mIsMetric;
    }
}
