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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;

import java.io.IOException;

/** Manages a leveled amount. */
public class LeveledAmount {
    private static final String KEY_AMOUNT    = "amount";
    public static final  String KEY_PER_LEVEL = "per_level";
    public static final  String KEY_DECIMAL   = "decimal";

    private double  mAmount;
    private int     mLevel;
    private boolean mPerLevel;
    private boolean mDecimal;

    /**
     * Creates a new leveled amount.
     *
     * @param amount The initial amount.
     */
    public LeveledAmount(double amount) {
        mPerLevel = false;
        mLevel = 0;
        mAmount = amount;
        mDecimal = true;
    }

    /**
     * Creates a new leveled amount.
     *
     * @param amount The initial amount.
     */
    public LeveledAmount(int amount) {
        this((double) amount);
        mDecimal = false;
    }

    /**
     * Creates a new leveled amount.
     *
     * @param other A LeveledAmount to clone.
     */
    public LeveledAmount(LeveledAmount other) {
        mPerLevel = other.mPerLevel;
        mLevel = other.mLevel;
        mAmount = other.mAmount;
        mDecimal = other.mDecimal;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof LeveledAmount amt) {
            return mPerLevel == amt.mPerLevel && mDecimal == amt.mDecimal && mLevel == amt.mLevel && mAmount == amt.mAmount;
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    public final void load(JsonMap m) {
        mDecimal = m.getBoolean(KEY_DECIMAL);
        mAmount = m.getDouble(KEY_AMOUNT);
        if (!mDecimal) {
            mAmount = Math.round(mAmount);
        }
        mPerLevel = m.getBoolean(KEY_PER_LEVEL);
    }

    public final void saveInline(JsonWriter w) throws IOException {
        if (mDecimal) {
            w.keyValue(KEY_AMOUNT, getAmount());
        } else {
            w.keyValue(KEY_AMOUNT, getIntegerAmount());
        }
        w.keyValueNot(KEY_DECIMAL, mDecimal, false);
        w.keyValueNot(KEY_PER_LEVEL, mPerLevel, false);
    }

    /** @return Whether the amount should be applied per level. */
    public boolean isPerLevel() {
        return mPerLevel;
    }

    /** @param perLevel Whether the amount should be applied per level. */
    public void setPerLevel(boolean perLevel) {
        mPerLevel = perLevel;
    }

    /** @return The current level to use. */
    public int getLevel() {
        return mLevel;
    }

    /** @param level The current level to use. */
    public void setLevel(int level) {
        mLevel = level;
    }

    /** @return Whether this is a decimal value. */
    public boolean isDecimal() {
        return mDecimal;
    }

    /** @param decimal Whether this is a decimal value. */
    public void setDecimal(boolean decimal) {
        mDecimal = decimal;
        if (!mDecimal) {
            mAmount = Math.round(mAmount);
        }
    }

    /** @return The amount. */
    public double getAmount() {
        return mAmount;
    }

    /** @return The amount. */
    public int getIntegerAmount() {
        return (int) Math.round(mAmount);
    }

    /** @return The amount, as a {@link String}. */
    public String getAmountAsString() {
        return getFormattedAmount(I18n.text("level"));
    }

    /** @return The amount, as a {@link String} of dice damage. */
    public String getAmountAsWeaponBonus() {
        return getFormattedAmount(I18n.text("die"));
    }

    private String getFormattedAmount(String what) {
        if (isPerLevel()) {
            String full;
            String perLevel;
            if (mDecimal) {
                full = Numbers.formatWithForcedSign(getAdjustedAmount());
                perLevel = Numbers.formatWithForcedSign(getAmount());
            } else {
                full = Numbers.formatWithForcedSign(getIntegerAdjustedAmount());
                perLevel = Numbers.formatWithForcedSign(getIntegerAmount());
            }
            return String.format(I18n.text("%s (%s per %s)"), full, perLevel, what);
        }
        if (mDecimal) {
            return Numbers.formatWithForcedSign(getAmount());
        }
        return Numbers.formatWithForcedSign(getIntegerAmount());
    }

    /** @param amount The amount. */
    public void setAmount(double amount) {
        mAmount = mDecimal ? amount : Math.round(amount);
    }

    /** @param amount The amount. */
    public void setAmount(int amount) {
        mAmount = amount;
    }

    /** @return The amount, adjusted for level, if requested. */
    public double getAdjustedAmount() {
        double amt = mAmount;
        if (isPerLevel()) {
            int level = getLevel();
            if (level < 0) {
                amt = 0;
            } else {
                amt *= level;
            }
        }
        return amt;
    }

    /** @return The amount, adjusted for level, if requested. */
    public int getIntegerAdjustedAmount() {
        int amt = getIntegerAmount();
        if (isPerLevel()) {
            int level = getLevel();
            if (level < 0) {
                amt = 0;
            } else {
                amt *= level;
            }
        }
        return amt;
    }
}
