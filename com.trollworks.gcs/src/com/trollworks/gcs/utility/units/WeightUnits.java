/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.units;

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.text.MessageFormat;

/** Common weight units. */
public enum WeightUnits implements Units {
    /** Ounces. */
    OZ(1.0 / 16.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Ounces");
        }
    },
    /** Pounds. */
    LB(1.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Pounds");
        }
    },
    /** Short Tons */
    TN(2000.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Short Tons");
        }
    },
    /** Long Tons */
    LT(2240.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Long Tons");
        }
    },
    /**
     * Metric Tons. Must come after Long Tons and Short Tons since it's abbreviation is a subset.
     */
    T(2205.0, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Metric Tons");
        }
    },
    /** Kilograms. */
    KG(2.205, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Kilograms");
        }
    },
    /** Grams. Must come after Kilograms since it's abbreviation is a subset. */
    G(0.002205, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Grams");
        }
    };

    private double  mFactor;
    private boolean mIsMetric;

    WeightUnits(double factor, boolean isMetric) {
        mFactor = factor;
        mIsMetric = isMetric;
    }

    @Override
    public double convert(Units units, double value) {
        return value * units.getFactor() / mFactor;
    }

    @Override
    public double normalize(double value) {
        return value * mFactor;
    }

    @Override
    public double getFactor() {
        return mFactor;
    }

    @Override
    public String format(double value, boolean localize) {
        String textValue = localize ? Numbers.format(value) : Double.toString(value);
        return MessageFormat.format("{0} {1}", Numbers.trimTrailingZeroes(textValue, localize), getAbbreviation());
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
    public String getDescription() {
        return String.format("%s (%s)", getLocalizedName(), getAbbreviation());
    }

    public boolean isMetric() {
        return mIsMetric;
    }
}
