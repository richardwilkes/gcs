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

package com.trollworks.gcs.body;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.util.Map;
import java.util.Objects;

public class HitLocation implements Cloneable, Comparable<HitLocation> {
    public static final  String           KEY_PREFIX      = "hit_location.";
    private static final String           KEY_ID          = "id";
    private static final String           KEY_CHOICE_NAME = "choice_name";
    private static final String           KEY_TABLE_NAME  = "table_name";
    private static final String           KEY_SLOTS       = "slots";
    private static final String           KEY_HIT_PENALTY = "hit_penalty";
    private static final String           KEY_DR_BONUS    = "dr_bonus";
    private static final String           KEY_DESCRIPTION = "description";
    private static final String           KEY_SUB_TABLE   = "sub_table";
    private              String           mID;
    private              String           mChoiceName;
    private              String           mTableName;
    private              int              mSlots;
    private              String           mRollRange;
    private              int              mHitPenalty;
    private              int              mDRBonus;
    private              String           mDescription;
    private              HitLocationTable mOwningTable;
    private              HitLocationTable mSubTable;

    public HitLocation(String id, String name, int slots, int hitPenalty, int drBonus, String description) {
        this(id, name, name, slots, hitPenalty, drBonus, description);
    }

    public HitLocation(String id, String choiceName, String tableName, int slots, int hitPenalty, int drBonus, String description) {
        setID(id);
        mChoiceName = choiceName;
        mTableName = tableName;
        mSlots = slots;
        mHitPenalty = hitPenalty;
        mDRBonus = drBonus;
        mDescription = description;
    }

    public HitLocation(JsonMap m) {
        setID(m.getString(KEY_ID));
        mChoiceName = m.getString(KEY_CHOICE_NAME);
        mTableName = m.getString(KEY_TABLE_NAME);
        mSlots = m.getInt(KEY_SLOTS);
        mHitPenalty = m.getInt(KEY_HIT_PENALTY);
        mDRBonus = m.getInt(KEY_DR_BONUS);
        mDescription = m.getString(KEY_DESCRIPTION);
        if (m.has(KEY_SUB_TABLE)) {
            setSubTable(new HitLocationTable(m.getMap(KEY_SUB_TABLE)));
        }
    }

    public void toJSON(JsonWriter w, GURPSCharacter character) throws IOException {
        w.startMap();
        w.keyValue(KEY_ID, mID);
        w.keyValue(KEY_CHOICE_NAME, mChoiceName);
        w.keyValue(KEY_TABLE_NAME, mTableName);
        w.keyValue(KEY_SLOTS, mSlots);
        w.keyValue(KEY_HIT_PENALTY, mHitPenalty);
        w.keyValue(KEY_DR_BONUS, mDRBonus);
        w.keyValue(KEY_DESCRIPTION, mDescription);
        if (mSubTable != null) {
            w.key(KEY_SUB_TABLE);
            mSubTable.toJSON(w, character);
        }

        // Emit the calculated values for third parties
        w.key("calc");
        w.startMap();
        w.keyValue("roll_range", getRollRange());
        if (character != null) {
            w.keyValue("dr", getDR(character, null));
        }
        w.endMap();

        w.endMap();
    }

    public String getID() {
        return mID;
    }

    public void setID(String id) {
        mID = com.trollworks.gcs.utility.ID.sanitize(id, null, false);
    }

    public String getChoiceName() {
        return mChoiceName;
    }

    public void setChoiceName(String name) {
        mChoiceName = name;
    }

    public String getTableName() {
        return mTableName;
    }

    public void setTableName(String name) {
        mTableName = name;
    }

    public int getSlots() {
        return mSlots;
    }

    public void setSlots(int slots) {
        mSlots = slots;
    }

    public String getRollRange() {
        return mRollRange;
    }

    public void setRollRange(String rollRange) {
        mRollRange = rollRange;
    }

    public int getHitPenalty() {
        return mHitPenalty;
    }

    public void setHitPenalty(int penalty) {
        mHitPenalty = penalty;
    }

    public int getDR(GURPSCharacter character, StringBuilder toolTip) {
        if (mDRBonus != 0 && toolTip != null) {
            toolTip.append("\n").append(mChoiceName).append(" [").append(Numbers.formatWithForcedSign(mDRBonus)).append("]");
        }
        int dr = character.getIntegerBonusFor(KEY_PREFIX + mID, toolTip) + mDRBonus;
        if (mOwningTable != null) {
            HitLocation owningLocation = mOwningTable.getOwningLocation();
            if (owningLocation != null) {
                dr += owningLocation.getDR(character, toolTip);
            }
        }
        return dr;
    }

    public int getDRBonus() {
        return mDRBonus;
    }

    public void setDRBonus(int bonus) {
        mDRBonus = bonus;
    }

    public String getDescription() {
        return mDescription;
    }

    public void setDescription(String description) {
        mDescription = description;
    }

    public HitLocationTable getOwningTable() {
        return mOwningTable;
    }

    public void setOwningTable(HitLocationTable table) {
        mOwningTable = table;
    }

    public HitLocationTable getSubTable() {
        return mSubTable;
    }

    public void setSubTable(HitLocationTable table) {
        if (table == null && mSubTable != null) {
            mSubTable.setOwningLocation(null);
        }
        mSubTable = table;
        if (mSubTable != null) {
            mSubTable.setOwningLocation(this);
        }
    }

    protected void populateMap(Map<String, HitLocation> map) {
        map.put(mID, this);
        if (mSubTable != null) {
            mSubTable.populateMap(map);
        }
    }

    @SuppressWarnings("MethodDoesntCallSuperMethod")
    @Override
    public HitLocation clone() {
        HitLocation other = new HitLocation(mID, mChoiceName, mTableName, mSlots, mHitPenalty, mDRBonus, mDescription);
        other.mRollRange = mRollRange;
        if (mSubTable != null) {
            other.setSubTable(mSubTable.clone());
        }
        return other;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        HitLocation that = (HitLocation) other;
        if (mSlots != that.mSlots) {
            return false;
        }
        if (mHitPenalty != that.mHitPenalty) {
            return false;
        }
        if (mDRBonus != that.mDRBonus) {
            return false;
        }
        if (!mID.equals(that.mID)) {
            return false;
        }
        if (!mChoiceName.equals(that.mChoiceName)) {
            return false;
        }
        if (!mTableName.equals(that.mTableName)) {
            return false;
        }
        if (!mDescription.equals(that.mDescription)) {
            return false;
        }
        return Objects.equals(mSubTable, that.mSubTable);
    }

    @Override
    public int hashCode() {
        int result = mID.hashCode();
        result = 31 * result + mChoiceName.hashCode();
        result = 31 * result + mTableName.hashCode();
        result = 31 * result + mSlots;
        result = 31 * result + mHitPenalty;
        result = 31 * result + mDRBonus;
        result = 31 * result + mDescription.hashCode();
        result = 31 * result + (mSubTable != null ? mSubTable.hashCode() : 0);
        return result;
    }

    @Override
    public int compareTo(HitLocation other) {
        int result = NumericComparator.caselessCompareStrings(mChoiceName, other.mChoiceName);
        if (result == 0) {
            result = NumericComparator.caselessCompareStrings(mID, other.mID);
        }
        return result;
    }

    @Override
    public String toString() {
        return mChoiceName;
    }
}
