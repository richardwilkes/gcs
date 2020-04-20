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

import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.I18n;

import java.text.MessageFormat;

/** Common length units. */
public enum LengthUnits implements Units {
    /** Points (1/72 of an inch). */
    PT(1.0 / 72.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Points");
        }
    },
    /** Inches. */
    IN(1.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Inches");
        }
    },
    /** Feet. */
    FT(12.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Feet");
        }
    },
    /** Feet and Inches */
    FT_IN(1.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Feet & Inches");
        }

        @Override
        public String getDescription() {
            return I18n.Text("Feet (') & Inches (\")");
        }

        @Override
        public String format(double value, boolean localize) {
            int feet = (int) (Math.floor(value) / 12);
            value -= 12.0 * feet;
            if (feet > 0) {
                String buffer = formatNumber(feet, localize) + '\'';
                if (value > 0) {
                    return buffer + ' ' + formatNumber(value, localize) + '"';
                }
                return buffer;
            }
            return formatNumber(value, localize) + '"';
        }

        private String formatNumber(double value, boolean localize) {
            return Numbers.trimTrailingZeroes(localize ? Numbers.format(value) : Double.toString(value), localize);
        }
    },
    /** Yards. */
    YD(36.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Yards");
        }
    },
    /** Miles. */
    MI(5280.0 * 12.0, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Miles");
        }
    },
    /** Millimeters. */
    MM(0.1 / 2.54, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Millimeters");
        }
    },
    /** Centimeters. */
    CM(1.0 / 2.54, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Centimeters");
        }
    },
    /** Kilometers. */
    KM(100000.0 / 2.54, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Kilometers");
        }
    },
    /** Meters. Must be after all the other 'meter' types. */
    M(100.0 / 2.54, true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Meters");
        }
    };

    private double  mFactor;
    private boolean mIsMetric;

    LengthUnits(double factor, boolean isMetric) {
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
    public LengthUnits[] getCompatibleUnits() {
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
