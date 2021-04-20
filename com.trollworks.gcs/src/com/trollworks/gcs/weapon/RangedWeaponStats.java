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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;

import java.io.IOException;

/** The stats for a ranged weapon. */
public class RangedWeaponStats extends WeaponStats {
    public static final  String KEY_ROOT         = "ranged_weapon";
    private static final String KEY_ACCURACY     = "accuracy";
    private static final String KEY_RANGE        = "range";
    private static final String KEY_RATE_OF_FIRE = "rate_of_fire";
    private static final String KEY_SHOTS        = "shots";
    private static final String KEY_BULK         = "bulk";
    private static final String KEY_RECOIL       = "recoil";

    private String mAccuracy;
    private String mRange;
    private String mRateOfFire;
    private String mShots;
    private String mBulk;
    private String mRecoil;

    /**
     * Creates a new {@link RangedWeaponStats}.
     *
     * @param owner The owning piece of equipment or advantage.
     */
    public RangedWeaponStats(ListRow owner) {
        super(owner);
    }

    /**
     * Creates a clone of the specified {@link RangedWeaponStats}.
     *
     * @param owner The owning piece of equipment or advantage.
     * @param other The {@link RangedWeaponStats} to clone.
     */
    public RangedWeaponStats(ListRow owner, RangedWeaponStats other) {
        super(owner, other);
        mAccuracy = other.mAccuracy;
        mRange = other.mRange;
        mRateOfFire = other.mRateOfFire;
        mShots = other.mShots;
        mBulk = other.mBulk;
        mRecoil = other.mRecoil;
    }

    /**
     * Creates a {@link RangedWeaponStats}.
     *
     * @param owner The owning piece of equipment or advantage.
     * @param m     The {@link JsonMap} to load from.
     */
    public RangedWeaponStats(ListRow owner, JsonMap m) throws IOException {
        super(owner, m);
    }

    @Override
    public WeaponStats clone(ListRow owner) {
        return new RangedWeaponStats(owner, this);
    }

    @Override
    protected void initialize() {
        mAccuracy = "";
        mRange = "";
        mRateOfFire = "";
        mShots = "";
        mBulk = "";
        mRecoil = "";
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        mAccuracy = m.getString(KEY_ACCURACY);
        mRange = m.getString(KEY_RANGE);
        mRateOfFire = m.getString(KEY_RATE_OF_FIRE);
        mShots = m.getString(KEY_SHOTS);
        mBulk = m.getString(KEY_BULK);
        mRecoil = m.getString(KEY_RECOIL);
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.keyValueNot(KEY_ACCURACY, mAccuracy, "");
        w.keyValueNot(KEY_RANGE, mRange, "");
        w.keyValueNot(KEY_RATE_OF_FIRE, mRateOfFire, "");
        w.keyValueNot(KEY_SHOTS, mShots, "");
        w.keyValueNot(KEY_BULK, mBulk, "");
        w.keyValueNot(KEY_RECOIL, mRecoil, "");
    }

    /** @return The accuracy. */
    public String getAccuracy() {
        return mAccuracy;
    }

    /**
     * Sets the value of accuracy.
     *
     * @param accuracy The value to set.
     */
    public void setAccuracy(String accuracy) {
        accuracy = sanitize(accuracy);
        if (!mAccuracy.equals(accuracy)) {
            mAccuracy = accuracy;
            notifyOfChange();
        }
    }

    /** @return The bulk. */
    public String getBulk() {
        return mBulk;
    }

    /**
     * Sets the value of bulk.
     *
     * @param bulk The value to set.
     */
    public void setBulk(String bulk) {
        bulk = sanitize(bulk);
        if (!mBulk.equals(bulk)) {
            mBulk = bulk;
            notifyOfChange();
        }
    }

    /** @return The range. */
    public String getRange() {
        return mRange;
    }

    /** @return The range, fully resolved for the user's ST, if possible. */
    public String getResolvedRange() {
        DataFile df    = getOwner().getDataFile();
        String   range = mRange;

        if (df instanceof GURPSCharacter) {
            int    strength = ((GURPSCharacter) df).getAttributeValue("st");
            String savedRange;

            do {
                savedRange = range;
                range = resolveRange(range, strength);
            } while (!savedRange.equals(range));
        }
        return range;
    }

    private String resolveRange(String range, int strength) {
        int where = range.indexOf('x');

        if (where != -1) {
            int last = where + 1;
            int max  = range.length();

            last = skipSpaces(range, last);
            if (last < max) {
                double  value      = 0.0;
                char    ch         = range.charAt(last);
                boolean found      = false;
                double  multiplier = 1.0;

                while (multiplier == 1.0 && ch == '.' || ch >= '0' && ch <= '9') {
                    found = true;
                    if (ch == '.') {
                        multiplier = 0.1;
                    } else if (multiplier == 1.0) {
                        value *= 10.0;
                        value += ch - '0';
                    } else {
                        value += (ch - '0') * multiplier;
                        multiplier *= 0.1;
                    }
                    if (++last >= max) {
                        break;
                    }
                    ch = range.charAt(last);
                }
                if (found) {
                    StringBuilder buffer = new StringBuilder();
                    if (where > 0) {
                        buffer.append(range, 0, where);
                    }
                    strength *= value;
                    buffer.append(Numbers.format(strength));
                    if (last < max) {
                        buffer.append(range.substring(last));
                    }
                    return buffer.toString();
                }
            }
        }
        return range;
    }

    /**
     * Sets the value of range.
     *
     * @param range The value to set.
     */
    public void setRange(String range) {
        range = sanitize(range);
        if (!mRange.equals(range)) {
            mRange = range;
            notifyOfChange();
        }
    }

    /** @return The rate of fire. */
    public String getRateOfFire() {
        return mRateOfFire;
    }

    /**
     * Sets the value of rate of fire.
     *
     * @param rateOfFire The value to set.
     */
    public void setRateOfFire(String rateOfFire) {
        rateOfFire = sanitize(rateOfFire);
        if (!mRateOfFire.equals(rateOfFire)) {
            mRateOfFire = rateOfFire;
            notifyOfChange();
        }
    }

    /** @return The recoil. */
    public String getRecoil() {
        return mRecoil;
    }

    /**
     * Sets the value of recoil.
     *
     * @param recoil The value to set.
     */
    public void setRecoil(String recoil) {
        recoil = sanitize(recoil);
        if (!mRecoil.equals(recoil)) {
            mRecoil = recoil;
            notifyOfChange();
        }
    }

    /** @return The shots. */
    public String getShots() {
        return mShots;
    }

    /**
     * Sets the value of shots.
     *
     * @param shots The value to set.
     */
    public void setShots(String shots) {
        shots = sanitize(shots);
        if (!mShots.equals(shots)) {
            mShots = shots;
            notifyOfChange();
        }
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof RangedWeaponStats && super.equals(obj)) {
            RangedWeaponStats rws = (RangedWeaponStats) obj;
            return mAccuracy.equals(rws.mAccuracy) && mRange.equals(rws.mRange) && mRateOfFire.equals(rws.mRateOfFire) && mShots.equals(rws.mShots) && mBulk.equals(rws.mBulk) && mRecoil.equals(rws.mRecoil);
        }
        return false;
    }
}
