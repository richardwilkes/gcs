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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.CharacterVariableResolver;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

public class Attribute {
    public static final  String ID_ATTR_PREFIX = "attr.";
    private static final String KEY_ATTR_ID    = "attr_id";
    private static final String KEY_ADJ        = "adj";

    private String mAttrDefID;
    private double mAdjustment;
    private double mBonus;
    private int    mCostReduction;

    public Attribute(String attrDefID) {
        mAttrDefID = attrDefID;
    }

    public Attribute(JsonMap m) {
        mAttrDefID = m.getString(KEY_ATTR_ID);
        mAdjustment = m.getDouble(KEY_ADJ);
    }

    public void initTo(double adjustment) {
        mAdjustment = adjustment;
    }

    public String getAttrID() {
        return ID_ATTR_PREFIX + mAttrDefID;
    }

    public String getAttrDefID() {
        return mAttrDefID;
    }

    public AttributeDef getAttrDef(GURPSCharacter character) {
        return character.getSettings().getAttributes().get(mAttrDefID);
    }

    public int getPointCost(GURPSCharacter character) {
        AttributeDef def = getAttrDef(character);
        if (def != null) {
            return def.computeCost(character, mAdjustment, character.getProfile().getSizeModifier(), mCostReduction);
        }
        return 0;
    }

    public int getIntValue(GURPSCharacter character) {
        return (int) Math.floor(getDoubleValue(character));
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

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_ATTR_ID, mAttrDefID);
        w.keyValue(KEY_ADJ, mAdjustment);
        w.endMap();
    }
}
