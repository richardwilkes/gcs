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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** Manages a leveled amount. */
public class LeveledAmount {
    /** The "per level" attribute. */
    public static final String  ATTRIBUTE_PER_LEVEL = "per_level";
    private static final String KEY_AMOUNT = "amount";
    private             boolean mPerLevel;
    private             int     mLevel;
    private             double  mAmount;
    private             boolean mInteger;

    /**
     * Creates a new leveled amount.
     *
     * @param amount The initial amount.
     */
    public LeveledAmount(double amount) {
        mPerLevel = false;
        mLevel = 0;
        mAmount = amount;
        mInteger = false;
    }

    /**
     * Creates a new leveled amount.
     *
     * @param amount The initial amount.
     */
    public LeveledAmount(int amount) {
        this((double) amount);
        mInteger = true;
    }

    /**
     * Creates a new leveled amount.
     *
     * @param other A {@link LeveledAmount} to clone.
     */
    public LeveledAmount(LeveledAmount other) {
        mPerLevel = other.mPerLevel;
        mLevel = other.mLevel;
        mAmount = other.mAmount;
        mInteger = other.mInteger;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof LeveledAmount) {
            LeveledAmount amt = (LeveledAmount) obj;
            return mPerLevel == amt.mPerLevel && mInteger == amt.mInteger && mLevel == amt.mLevel && mAmount == amt.mAmount;
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    /** @param reader The reader to load data from. */
    public void load(XMLReader reader) throws IOException {
        mPerLevel = reader.isAttributeSet(ATTRIBUTE_PER_LEVEL);
        mAmount = reader.readDouble(0.0);
    }

    public final void load(JsonMap m) throws IOException {
        mAmount = m.getDouble(KEY_AMOUNT);
        if (mInteger) {
            mAmount = Math.round(mAmount);
        }
        mPerLevel = m.getBoolean(ATTRIBUTE_PER_LEVEL);
    }

    public final void saveInline(JsonWriter w) throws IOException {
        if (mInteger) {
            w.keyValue(KEY_AMOUNT, getIntegerAmount());
        } else {
            w.keyValue(KEY_AMOUNT, getAmount());
        }
        w.keyValueNot(ATTRIBUTE_PER_LEVEL, mPerLevel, false);
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

    /** @return Whether this is an integer representation only. */
    public boolean isIntegerOnly() {
        return mInteger;
    }

    /** @param integerOnly Whether this is an integer representation only. */
    public void setIntegerOnly(boolean integerOnly) {
        mInteger = integerOnly;
        if (mInteger) {
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
        return getFormattedAmount(I18n.Text("level"));
    }

    /** @return The amount, as a {@link String} of dice damage. */
    public String getAmountAsWeaponBonus() {
        return getFormattedAmount(I18n.Text("die"));
    }

    private String getFormattedAmount(String what) {
        if (isPerLevel()) {
            String full;
            String perLevel;
            if (mInteger) {
                full = Numbers.formatWithForcedSign(getIntegerAdjustedAmount());
                perLevel = Numbers.formatWithForcedSign(getIntegerAmount());
            } else {
                full = Numbers.formatWithForcedSign(getAdjustedAmount());
                perLevel = Numbers.formatWithForcedSign(getAmount());
            }
            return String.format(I18n.Text("%s (%s per %s)"), full, perLevel, what);
        }
        if (mInteger) {
            return Numbers.formatWithForcedSign(getIntegerAmount());
        }
        return Numbers.formatWithForcedSign(getAmount());
    }

    /** @param amount The amount. */
    public void setAmount(double amount) {
        mAmount = mInteger ? Math.round(amount) : amount;
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
