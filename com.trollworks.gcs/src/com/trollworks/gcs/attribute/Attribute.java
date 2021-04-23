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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.character.CharacterVariableResolver;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.List;

public class Attribute {
    public static final  String ID_ATTR_PREFIX = "attr.";
    private static final String KEY_ATTR_ID    = "attr_id";
    private static final String KEY_ADJ        = "adj";
    private static final String KEY_DAMAGE     = "damage";

    private String mAttrID;
    private double mAdjustment;
    private double mBonus;
    private int    mCostReduction;
    private int    mDamage;

    public Attribute(String attrID) {
        mAttrID = attrID;
    }

    public Attribute(JsonMap m) {
        mAttrID = m.getString(KEY_ATTR_ID);
        mAdjustment = m.getDouble(KEY_ADJ);
        mDamage = m.getInt(KEY_DAMAGE);
    }

    public void initTo(double adjustment, int damage) {
        mAdjustment = adjustment;
        mDamage = damage;
    }

    public String getID() {
        return mAttrID;
    }

    public AttributeDef getAttrDef(GURPSCharacter character) {
        return character.getSettings().getAttributes().get(mAttrID);
    }

    public PoolThreshold getCurrentThreshold(GURPSCharacter character) {
        AttributeDef def = getAttrDef(character);
        if (def != null) {
            List<PoolThreshold> thresholds = def.getThresholds();
            if (thresholds != null) {
                int max = getIntValue(character);
                int cur = max - mDamage;
                for (PoolThreshold threshold : thresholds) {
                    if (cur <= threshold.threshold(max)) {
                        return threshold;
                    }
                }
            }
        }
        return null;
    }

    public int getPointCost(GURPSCharacter character) {
        AttributeDef def = getAttrDef(character);
        if (def != null) {
            return def.computeCost(character, mAdjustment, character.getProfile().getSizeModifier(), mCostReduction);
        }
        return 0;
    }

    public int getCurrentIntValue(GURPSCharacter character) {
        return getIntValue(character) - mDamage;
    }

    public int getIntValue(GURPSCharacter character) {
        return (int) getDoubleValue(character);
    }

    public double getDoubleValue(GURPSCharacter character) {
        AttributeDef def = getAttrDef(character);
        if (def != null) {
            return def.getBaseValue(new CharacterVariableResolver(character)) + mAdjustment + mBonus;
        }
        return 0;
    }

    public void setIntValue(GURPSCharacter character, int value) {
        int old = getIntValue(character);
        if (old != value) {
            AttributeDef def = getAttrDef(character);
            if (def != null) {
                character.postUndoEdit(String.format(I18n.Text("%s Change"), def.getName()), (c, v) -> setIntValue(c, ((Integer) v).intValue()), Integer.valueOf(old), Integer.valueOf(value));
                mAdjustment = value - (def.getBaseValue(new CharacterVariableResolver(character)) + mBonus);
                character.notifyOfChange();
            }
        }
    }

    public void setDoubleValue(GURPSCharacter character, double value) {
        double old = getDoubleValue(character);
        if (old != value) {
            AttributeDef def = getAttrDef(character);
            if (def != null) {
                character.postUndoEdit(String.format(I18n.Text("%s Change"), def.getName()), (c, v) -> setDoubleValue(c, ((Double) v).doubleValue()), Double.valueOf(old), Double.valueOf(value));
                mAdjustment = value - (def.getBaseValue(new CharacterVariableResolver(character)) + mBonus);
                character.notifyOfChange();
            }
        }
    }

    public void setBonus(GURPSCharacter character, double bonus) {
        if (mBonus != bonus) {
            mBonus = bonus;
            character.notifyOfChange();
        }
    }

    public void setCostReduction(GURPSCharacter character, int reductionPercentage) {
        if (mCostReduction != reductionPercentage) {
            mCostReduction = reductionPercentage;
            character.notifyOfChange();
        }
    }

    public int getDamage() {
        return mDamage;
    }

    public void setDamage(GURPSCharacter character, int damage) {
        if (mDamage != damage) {
            AttributeDef def = getAttrDef(character);
            if (def != null) {
                character.postUndoEdit(String.format(I18n.Text("Current %s Change"), def.getName()), (c, v) -> setDamage(c, ((Integer) v).intValue()), Integer.valueOf(mDamage), Integer.valueOf(damage));
                mDamage = damage;
                character.notifyOfChange();
            }
        }
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_ATTR_ID, mAttrID);
        w.keyValue(KEY_ADJ, mAdjustment);
        w.keyValue(KEY_DAMAGE, mDamage);
        w.endMap();
    }
}
