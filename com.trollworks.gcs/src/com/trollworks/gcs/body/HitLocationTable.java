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
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class HitLocationTable implements Cloneable {
    private static final String                        KEY_ID        = "id";
    private static final String                        KEY_NAME      = "name";
    private static final String                        KEY_ROLL      = "roll";
    private static final String                        KEY_LOCATIONS = "locations";
    private static       List<HitLocationTable>        STD_TABLES_LIST;
    private static       Map<String, HitLocationTable> STD_TABLES_MAP;
    private              String                        mID;
    private              String                        mName;
    private              Dice                          mRoll;
    private              List<HitLocation>             mLocations;
    private              HitLocation                   mOwningLocation;
    private              Map<String, HitLocation>      mLocationLookup;

    public HitLocationTable(String id, String name, Dice roll) {
        setID(id);
        mName = name;
        mRoll = roll.clone();
        mLocations = new ArrayList<>();
    }

    public HitLocationTable(JsonMap m) {
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

    public static final synchronized List<HitLocationTable> getStdTables() {
        if (STD_TABLES_LIST == null) {
            STD_TABLES_LIST = Collections.unmodifiableList(createStdTables());
        }
        return STD_TABLES_LIST;
    }

    public static final synchronized HitLocationTable lookupStdTable(String id) {
        if (STD_TABLES_MAP == null) {
            STD_TABLES_MAP = new HashMap<>();
            for (HitLocationTable table : getStdTables()) {
                STD_TABLES_MAP.put(table.getID(), table);
            }
        }
        return STD_TABLES_MAP.get(id);
    }

    private static List<HitLocationTable> createStdTables() {
        List<HitLocationTable> tables = new ArrayList<>();
        tables.add(createArachnoidTable());
        tables.add(createAvianTable());
        tables.add(createCancroidTable());
        tables.add(createCentaurTable());
        tables.add(createHexapodTable());
        tables.add(createWingedHexapodTable());
        tables.add(createHumanoidTable());
        tables.add(createIchthyoidTable());
        tables.add(createOctopodTable());
        tables.add(createQuadrupedTable());
        tables.add(createWingedQuadrupedTable());
        tables.add(createScorpionTable());
        tables.add(createSnakemenTable());
        tables.add(createSquidTable());
        tables.add(createVeriformTable());
        tables.add(createWingedVeriformTable());
        return tables;
    }

    private static HitLocationTable createArachnoidTable() {
        HitLocationTable table = new HitLocationTable("arachnoid", I18n.Text("Arachnoid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("brain", I18n.Text("Brain"), 2, -7, 1, getBrainDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Leg 1-2"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Leg 3-4"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Leg 5-6"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Leg 7-8"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createAvianTable() {
        HitLocationTable table = new HitLocationTable("avian", I18n.Text("Avian"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 2, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createCancroidTable() {
        HitLocationTable table = new HitLocationTable("cancroid", I18n.Text("Cancroid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), 4, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createCentaurTable() {
        HitLocationTable table = new HitLocationTable("centaur", I18n.Text("Centaur"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("hand", I18n.Text("Extremity"), 2, -4, 0, getHandDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createHexapodTable() {
        HitLocationTable table = new HitLocationTable("hexapod", I18n.Text("Hexapod"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Midleg"), 1, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Midleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createWingedHexapodTable() {
        HitLocationTable table = new HitLocationTable("hexapod_winged", I18n.Text("Hexapod, Winged"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Midleg"), 1, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 1, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Midleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    public static final HitLocationTable createHumanoidTable() {
        HitLocationTable table = new HitLocationTable("humanoid", I18n.Text("Humanoid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Right Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Right Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Left Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Left Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("hand", I18n.Text("Hand"), 1, -4, 0, getHandDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 1, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 2, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createIchthyoidTable() {
        HitLocationTable table = new HitLocationTable("ichthyoid", I18n.Text("Ichthyoid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -8, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("fin", I18n.Text("Fin"), 1, -4, 0, getFinDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 6, 0, 0, ""));
        table.addLocation(new HitLocation("fin", I18n.Text("Fin"), 4, -4, 0, getFinDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createOctopodTable() {
        HitLocationTable table = new HitLocationTable("octopod", I18n.Text("Octopod"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -8, 0, getEyeDescription()));
        table.addLocation(new HitLocation("brain", I18n.Text("Brain"), 2, -7, 1, getBrainDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Arm 1-2"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Arm 3-4"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Arm 5-6"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Arm 7-8"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createQuadrupedTable() {
        HitLocationTable table = new HitLocationTable("quadruped", I18n.Text("Quadruped"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createWingedQuadrupedTable() {
        HitLocationTable table = new HitLocationTable("quadruped_winged", I18n.Text("Quadruped, Winged"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 1, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }


    private static HitLocationTable createScorpionTable() {
        HitLocationTable table = new HitLocationTable("scorpion", I18n.Text("Scorpion"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 1, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), 4, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createSnakemenTable() {
        HitLocationTable table = new HitLocationTable("snakemen", I18n.Text("Snakemen"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("hand", I18n.Text("Hand"), 2, -4, 0, getHandDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createSquidTable() {
        HitLocationTable table = new HitLocationTable("squid", I18n.Text("Squid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -8, 0, getEyeDescription()));
        table.addLocation(new HitLocation("brain", I18n.Text("Brain"), 2, -7, 1, getBrainDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("hand", I18n.Text("Extremity"), 4, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createVeriformTable() {
        HitLocationTable table = new HitLocationTable("vermiform", I18n.Text("Vermiform"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 3, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 10, 0, 0, ""));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static HitLocationTable createWingedVeriformTable() {
        HitLocationTable table = new HitLocationTable("vermiform_winged", I18n.Text("Vermiform, Winged"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 3, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 6, 0, 0, ""));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 4, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        table.update();
        return table;
    }

    private static String getEyeDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Only impaling (imp), piercing (pi-, pi, pi+, pi++), and tight-beam burning (burn) attacks can target the eye – and only from the front or sides. Injury over HP÷10 blinds the eye. Otherwise, treat as skull, but without the extra DR!");
    }

    private static String getSkullDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Wounding modifier is x4. Knockdown rolls are at -10. Critical hits use the Critical Head Blow Table (B556). Exception: These special effects do not apply to toxic (tox) damage.");
    }

    private static String getFaceDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Jaw, cheeks, nose, ears, etc. If the target has an open-faced helmet, ignore its DR. Knockdown rolls are at -5. Critical hits use the Critical Head Blow Table (B556). Corrosion (cor) damage gets a x1½ wounding modifier, and if it inflicts a major wound, it also blinds one eye (both eyes on damage over full HP). Random attacks from behind hit the skull instead.");
    }

    private static String getLimbDescription() {
        return I18n.Text("Reduce the wounding multiplier of large piercing (pi+), huge piercing (pi++), and impaling (imp) damage to x1. Any major wound (loss of over ½ HP from one blow) cripples the limb. Damage beyond that threshold is lost.");
    }

    private static String getArmDescription() {
        return String.format(I18n.Text("%s If holding a shield, double the penalty to hit: -4 for shield arm instead of -2."), getLimbDescription());
    }

    private static String getGroinDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Human males and the males of similar species suffer double shock from crushing (cr) damage, and get -5 to knockdown rolls. Otherwise, treat as a torso hit.");
    }

    private static String getHandDescription() {
        return String.format(I18n.Text("If holding a shield, double the penalty to hit: -8 for shield hand instead of -4. %s"), getExtremityDescription());
    }

    private static String getExtremityDescription() {
        return I18n.Text("Reduce the wounding multiplier of large piercing (pi+), huge piercing (pi++), and impaling (imp) damage to x1. Any major wound (loss of over ⅓ HP from one blow) cripples the extremity. Damage beyond that threshold is lost.");
    }

    private static String getNeckDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Neck and throat. Increase the wounding multiplier of crushing (cr) and corrosion (cor) attacks to x1½, and that of cutting (cut) damage to x2. At the GM’s option, anyone killed by a cutting (cut) blow to the neck is decapitated!");
    }

    private static String getVitalsDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Heart, lungs, kidneys, etc. Increase the wounding modifier for an impaling (imp) or any piercing (pi-, pi, pi+, pi++) attack to x3. Increase the wounding modifier for a tight-beam burning (burn) attack to x2. Other attacks cannot target the vitals.");
    }

    private static String getTailDescription() {
        return I18n.Text("If a tail counts as an Extra Arm or a Striker, or is a fish tail, treat it as a limb (arm, leg) for crippling purposes; otherwise, treat it as an extremity (hand, foot). A crippled tail affects balance. For a ground creature, this gives -1 DX. For a swimmer or flyer, this gives -2 DX and halves Move. If the creature has no tail, or a very short one (like a rabbit), treat as torso.");
    }

    private static String getWingDescription() {
        return String.format(I18n.Text("%s A flyer with a crippled wing cannot fly."), getLimbDescription());
    }

    private static String getBrainDescription() {
        return I18n.Text("An attack that misses by 1 hits the torso instead. Wounding modifier is x4. Knockdown rolls are at -10. Critical hits use the Critical Head Blow Table (B556). Exception: These special effects do not apply to toxic (tox) damage.");
    }

    private static String getFinDescription() {
        return String.format(I18n.Text("%s A crippled fin affects balance: -3 DX."), getExtremityDescription());
    }
}
