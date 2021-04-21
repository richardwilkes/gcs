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

package com.trollworks.gcs.pointpool;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.CharacterVariableResolver;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

public class PointPool {
    public static final  String ID_POOL_PREFIX = "pool.";
    private static final String KEY_POOL_ID    = "pool_id";
    private static final String KEY_ADJ        = "adj";
    private static final String KEY_DAMAGE     = "damage";

    private String mPoolDefID;
    private int    mAdjustment;
    private int    mDamage;
    private int    mBonus;

    public PointPool(String poolDefID) {
        mPoolDefID = poolDefID;
    }

    public PointPool(JsonMap m) {
        mPoolDefID = m.getString(KEY_POOL_ID);
        mAdjustment = m.getInt(KEY_ADJ);
        mDamage = m.getInt(KEY_DAMAGE);
    }

    public void initTo(int adjustment, int damage) {
        mAdjustment = adjustment;
        mDamage = damage;
    }

    public String getPoolID() {
        return ID_POOL_PREFIX + mPoolDefID;
    }

    public String getPoolDefID() {
        return mPoolDefID;
    }

    public PointPoolDef getPoolDef(GURPSCharacter character) {
        return character.getSettings().getPointPools().get(mPoolDefID);
    }

    public PoolThreshold getCurrentThreshold(GURPSCharacter character) {
        PointPoolDef pool = getPoolDef(character);
        if (pool != null) {
            int max = getMaximum(character);
            int cur = max - mDamage;
            for (PoolThreshold threshold : pool.getThresholds()) {
                if (cur <= threshold.threshold(max)) {
                    return threshold;
                }
            }
        }
        return null;
    }

    public int getPointCost(GURPSCharacter character) {
        PointPoolDef pool = getPoolDef(character);
        if (pool != null) {
            return pool.computeCost(character, mAdjustment, character.getProfile().getSizeModifier());
        }
        return 0;
    }

    public int getCurrent(GURPSCharacter character) {
        return getMaximum(character) - mDamage;
    }

    public int getMaximum(GURPSCharacter character) {
        PointPoolDef pool = getPoolDef(character);
        if (pool != null) {
            return pool.getBaseValue(new CharacterVariableResolver(character)) + mAdjustment + mBonus;
        }
        return 0;
    }

    public void setMaximum(GURPSCharacter character, int value) {
        int old = getMaximum(character);
        if (old != value) {
            PointPoolDef pool = getPoolDef(character);
            if (pool != null) {
                character.postUndoEdit(String.format(I18n.Text("%s Change"), pool.getName()), (c, v) -> setMaximum(c, ((Integer) v).intValue()), Integer.valueOf(old), Integer.valueOf(value));
                mAdjustment = value - (pool.getBaseValue(new CharacterVariableResolver(character)) + mBonus);
                character.notifyOfChange();
            }
        }
    }

    public int getDamage() {
        return mDamage;
    }

    public void setDamage(GURPSCharacter character, int damage) {
        if (mDamage != damage) {
            PointPoolDef pool = getPoolDef(character);
            if (pool != null) {
                character.postUndoEdit(String.format(I18n.Text("Current %s Change"), pool.getName()), (c, v) -> setDamage(c, ((Integer) v).intValue()), Integer.valueOf(mDamage), Integer.valueOf(damage));
                mDamage = damage;
                character.notifyOfChange();
            }
        }
    }

    public int getBonus() {
        return mBonus;
    }

    public void setBonus(GURPSCharacter character, int bonus) {
        if (mBonus != bonus) {
            mBonus = bonus;
            character.notifyOfChange();
        }
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_POOL_ID, mPoolDefID);
        w.keyValue(KEY_ADJ, mAdjustment);
        w.keyValue(KEY_DAMAGE, mDamage);
        w.endMap();
    }
}
