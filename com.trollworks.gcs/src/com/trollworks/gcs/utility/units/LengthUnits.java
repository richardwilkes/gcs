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

/** Common length units. */
public enum LengthUnits implements Units {
    /** Points (1/72 of an inch). */
    PT(Fixed6.ONE.div(new Fixed6(72)), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Points");
        }
    },
    /** Inches. */
    IN(Fixed6.ONE, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Inches");
        }
    },
    /** Feet. */
    FT(new Fixed6(12), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Feet");
        }
    },
    /** Feet and Inches */
    FT_IN(Fixed6.ONE, false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Feet & Inches");
        }

        @Override
        public String toString() {
            return I18n.Text("Feet (') & Inches (\")");
        }

        @Override
        public String format(Fixed6 value, boolean localize) {
            Fixed6 twelve = new Fixed6(12);
            Fixed6 feet   = value.div(twelve).trunc();
            Fixed6 inches = value.sub(twelve.mul(feet));
            if (feet.greaterThan(Fixed6.ZERO)) {
                String buffer = formatNumber(feet, localize) + "'";
                if (inches.greaterThan(Fixed6.ZERO)) {
                    return buffer + ' ' + formatNumber(inches, localize) + '"';
                }
                return buffer;
            }
            return formatNumber(value, localize) + '"';
        }

        private String formatNumber(Fixed6 value, boolean localize) {
            return localize ? value.toLocalizedString() : value.toString();
        }
    },
    /** Yards. */
    YD(new Fixed6(36), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Yards");
        }
    },
    /** Miles. */
    MI(new Fixed6(5280).mul(new Fixed6(12)), false) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Miles");
        }
    },
    /** Millimeters. */
    MM(new Fixed6("0.1", Fixed6.ZERO, false).div(LengthValue.METRIC_CONVERSION_FACTOR), true) { // entered as text to ensure precision
        @Override
        public String getLocalizedName() {
            return I18n.Text("Millimeters");
        }
    },
    /** Centimeters. */
    CM(Fixed6.ONE.div(LengthValue.METRIC_CONVERSION_FACTOR), true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Centimeters");
        }
    },
    /** Kilometers. */
    KM(new Fixed6(100000).div(LengthValue.METRIC_CONVERSION_FACTOR), true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Kilometers");
        }
    },
    /** Meters. Must be after all the other 'meter' types. */
    M(new Fixed6(100).div(LengthValue.METRIC_CONVERSION_FACTOR), true) {
        @Override
        public String getLocalizedName() {
            return I18n.Text("Meters");
        }
    };

    private Fixed6  mFactor;
    private boolean mIsMetric;

    LengthUnits(Fixed6 factor, boolean isMetric) {
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
    public LengthUnits[] getCompatibleUnits() {
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
