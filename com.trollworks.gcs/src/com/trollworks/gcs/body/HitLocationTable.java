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
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class HitLocationTable implements Cloneable, Comparable<HitLocationTable> {
    private static final String                   KEY_ID        = "id";
    private static final String                   KEY_NAME      = "name";
    private static final String                   KEY_ROLL      = "roll";
    private static final String                   KEY_LOCATIONS = "locations";
    private              String                   mID;
    private              String                   mName;
    private              Dice                     mRoll;
    private              List<HitLocation>        mLocations;
    private              HitLocation              mOwningLocation;
    private              Map<String, HitLocation> mLocationLookup;

    public HitLocationTable(String id, String name, Dice roll) {
        setID(id);
        mName = name;
        mRoll = roll.clone();
        mLocations = new ArrayList<>();
    }

    public HitLocationTable(Path path) throws IOException {
        try (BufferedReader fileReader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m       = Json.asMap(Json.parse(fileReader));
            int     version = m.getInt(DataFile.VERSION);
            if (version > DataFile.CURRENT_VERSION) {
                throw VersionException.createTooNew();
            }
            if (!m.has(SheetSettings.KEY_HIT_LOCATIONS)) {
                throw new IOException("invalid data type");
            }
            loadJSON(m.getMap(SheetSettings.KEY_HIT_LOCATIONS));
        }
    }

    public HitLocationTable(JsonMap m) {
        loadJSON(m);
    }

    private void loadJSON(JsonMap m) {
        setID(m.getString(KEY_ID));
        mName = m.getString(KEY_NAME);
        mRoll = new Dice(m.getString(KEY_ROLL));
        mLocations = new ArrayList<>();
        JsonArray a    = m.getArray(KEY_LOCATIONS);
        int       size = a.size();
        for (int i = 0; i < size; i++) {
            addLocation(new HitLocation(a.getMap(i)));
        }
        update();
    }

    public void toJSON(JsonWriter w, GURPSCharacter character) throws IOException {
        w.startMap();
        w.keyValue(KEY_ID, mID);
        w.keyValue(KEY_NAME, mName);
        w.keyValue(KEY_ROLL, mRoll.toString(false));
        w.key(KEY_LOCATIONS);
        w.startArray();
        for (HitLocation location : mLocations) {
            location.toJSON(w, character);
        }
        w.endArray();
        w.endMap();
    }

    public void update() {
        updateRollRanges();
        mLocationLookup = new HashMap<>();
        populateMap(mLocationLookup);
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

    public void removeLocation(HitLocation location) {
        mLocations.remove(location);
        location.setOwningTable(null);
    }

    public HitLocation getOwningLocation() {
        return mOwningLocation;
    }

    public void setOwningLocation(HitLocation owningLocation) {
        mOwningLocation = owningLocation;
        if (mOwningLocation != null) {
            mID = "";
            mName = "";
        }
    }

    private void updateRollRanges() {
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
            HitLocationTable subTable = location.getSubTable();
            if (subTable != null) {
                subTable.updateRollRanges();
            }
            start += slots;
        }
    }

    protected void populateMap(Map<String, HitLocation> map) {
        for (HitLocation location : mLocations) {
            location.populateMap(map);
        }
    }

    public List<HitLocation> getUniqueHitLocations() {
        if (mLocationLookup == null) {
            update();
        }
        List<HitLocation> locations = new ArrayList<>(mLocationLookup.values());
        Collections.sort(locations);
        return locations;
    }

    public HitLocation lookupLocationByID(String id) {
        if (mLocationLookup == null) {
            update();
        }
        return mLocationLookup.get(id);
    }

    @SuppressWarnings("MethodDoesntCallSuperMethod")
    @Override
    public HitLocationTable clone() {
        HitLocationTable other = new HitLocationTable(mID, mName, mRoll);
        for (HitLocation location : mLocations) {
            other.addLocation(location.clone());
        }
        other.update();
        return other;
    }

    public void resetTo(HitLocationTable other) {
        mID = other.mID;
        mName = other.mName;
        mRoll = other.mRoll;
        mLocations = new ArrayList<>();
        for (HitLocation location : other.mLocations) {
            addLocation(location.clone());
        }
        update();
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (other == null || getClass() != other.getClass()) {
            return false;
        }
        HitLocationTable that = (HitLocationTable) other;
        if (!mID.equals(that.mID)) {
            return false;
        }
        if (!mName.equals(that.mName)) {
            return false;
        }
        if (!mRoll.equals(that.mRoll)) {
            return false;
        }
        return mLocations.equals(that.mLocations);
    }

    @Override
    public int hashCode() {
        int result = mID.hashCode();
        result = 31 * result + mName.hashCode();
        result = 31 * result + mRoll.hashCode();
        result = 31 * result + mLocations.hashCode();
        return result;
    }

    @Override
    public int compareTo(HitLocationTable other) {
        int result = NumericComparator.caselessCompareStrings(mName, other.mName);
        if (result == 0) {
            result = NumericComparator.caselessCompareStrings(mID, other.mID);
        }
        return result;
    }
}
