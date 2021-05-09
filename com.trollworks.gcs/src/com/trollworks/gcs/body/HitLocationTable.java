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

import com.trollworks.gcs.utility.Dice;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class HitLocationTable {
    private String            mID;
    private String            mName;
    private Dice              mRoll;
    private List<HitLocation> mLocations;
    private HitLocation       mOwningLocation;

    public HitLocationTable(String id, String name, Dice roll) {
        setID(id);
        mName = name;
        mRoll = roll.clone();
        mLocations = new ArrayList<>();
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

    public Dice getRoll() {
        return mRoll;
    }

    public void setRoll(Dice roll) {
        mRoll = roll;
    }

    public List<HitLocation> getLocations() {
        return mLocations;
    }

    public void addLocation(HitLocation location) {
        mLocations.add(location);
        location.setOwningTable(this);
    }

    public HitLocation getOwningLocation() {
        return mOwningLocation;
    }

    public void setOwningLocation(HitLocation owningLocation) {
        mOwningLocation = owningLocation;
    }

    public void updateRollRanges() {
        int start = mRoll.min(false);
        for (HitLocation location : mLocations) {
            String rollRange;
            int    slots = location.getSlots();
            switch (slots) {
            case 0 -> rollRange = "-";
            case 1 -> rollRange = Integer.toString(start);
            default -> rollRange = start + "-" + (start + slots - 1);
            }
            location.setRollRange(rollRange);
            HitLocationTable subTable = location.setSubTable();
            if (subTable != null) {
                subTable.updateRollRanges();
            }
            start += slots;
        }
    }

    public void populateMap(Map<String, HitLocation> map) {
        for (HitLocation location : mLocations) {
            location.populateMap(map);
        }
    }
}
