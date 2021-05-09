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

import java.util.Map;

public class HitLocation {
    private String           mID;
    private String           mName;
    private int              mSlots;
    private String           mRollRange;
    private int              mHitPenalty;
    private int              mDRBonus;
    private String           mDescription;
    private HitLocationTable mOwningTable;
    private HitLocationTable mSubTable;

    public HitLocation(String id, String name, int slots, int hitPenalty, int drBonus, String description) {
        setID(id);
        mName = name;
        mSlots = slots;
        mHitPenalty = hitPenalty;
        mDRBonus = drBonus;
        mDescription = description;
    }

    public String getID() {
        return mID;
    }

    public void setID(String id) {
        mID = com.trollworks.gcs.utility.ID.sanitize(id, null, false);
    }

    public String getName() {
        return mName;
    }

    public void setName(String name) {
        mName = name;
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

    public HitLocationTable setSubTable() {
        return mSubTable;
    }

    public void setSubTable(HitLocationTable table) {
        mSubTable = table;
        mSubTable.setOwningLocation(this);
    }

    public void populateMap(Map<String, HitLocation> map) {
        map.put(mID, this);
        if (mSubTable != null) {
            mSubTable.populateMap(map);
        }
    }
}
