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
import com.trollworks.gcs.utility.I18n;
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

    public static HitLocationTable createHumanoidTable() {
        HitLocationTable table = new HitLocationTable("humanoid", I18n.text("Humanoid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.text("Eyes"), 0, -9, 0, I18n.text("An attack that misses by 1 hits the torso instead. Only impaling (imp), piercing (pi-, pi, pi+, pi++), and tight-beam burning (burn) attacks can target the eye – and only from the front or sides. Injury over HP÷10 blinds the eye. Otherwise, treat as skull, but without the extra DR!")));
        table.addLocation(new HitLocation("skull", I18n.text("Skull"), 2, -7, 2, I18n.text("An attack that misses by 1 hits the torso instead. Wounding modifier is x4. Knockdown rolls are at -10. Critical hits use the Critical Head Blow Table (B556). Exception: These special effects do not apply to toxic (tox) damage.")));
        table.addLocation(new HitLocation("face", I18n.text("Face"), 1, -5, 0, I18n.text("An attack that misses by 1 hits the torso instead. Jaw, cheeks, nose, ears, etc. If the target has an open-faced helmet, ignore its DR. Knockdown rolls are at -5. Critical hits use the Critical Head Blow Table (B556). Corrosion (cor) damage gets a x1½ wounding modifier, and if it inflicts a major wound, it also blinds one eye (both eyes on damage over full HP). Random attacks from behind hit the skull instead.")));
        table.addLocation(new HitLocation("leg", I18n.text("Leg"), I18n.text("Right Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("arm", I18n.text("Arm"), I18n.text("Right Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.text("Groin"), 1, -3, 0, I18n.text("An attack that misses by 1 hits the torso instead. Human males and the males of similar species suffer double shock from crushing (cr) damage, and get -5 to knockdown rolls. Otherwise, treat as a torso hit.")));
        table.addLocation(new HitLocation("arm", I18n.text("Arm"), I18n.text("Left Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("leg", I18n.text("Leg"), I18n.text("Left Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("hand", I18n.text("Hand"), 1, -4, 0, String.format(I18n.text("If holding a shield, double the penalty to hit: -8 for shield hand instead of -4. %s"), getExtremityDescription())));
        table.addLocation(new HitLocation("foot", I18n.text("Foot"), 1, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("neck", I18n.text("Neck"), 2, -5, 0, I18n.text("An attack that misses by 1 hits the torso instead. Neck and throat. Increase the wounding multiplier of crushing (cr) and corrosion (cor) attacks to x1½, and that of cutting (cut) damage to x2. At the GM’s option, anyone killed by a cutting (cut) blow to the neck is decapitated!")));
        table.addLocation(new HitLocation("vitals", I18n.text("Vitals"), 0, -3, 0, I18n.text("An attack that misses by 1 hits the torso instead. Heart, lungs, kidneys, etc. Increase the wounding modifier for an impaling (imp) or any piercing (pi-, pi, pi+, pi++) attack to x3. Increase the wounding modifier for a tight-beam burning (burn) attack to x2. Other attacks cannot target the vitals.")));
        table.update();
        return table;
    }

    private static String getLimbDescription() {
        return I18n.text("Reduce the wounding multiplier of large piercing (pi+), huge piercing (pi++), and impaling (imp) damage to x1. Any major wound (loss of over ½ HP from one blow) cripples the limb. Damage beyond that threshold is lost.");
    }

    private static String getArmDescription() {
        return String.format(I18n.text("%s If holding a shield, double the penalty to hit: -4 for shield arm instead of -2."), getLimbDescription());
    }

    private static String getExtremityDescription() {
        return I18n.text("Reduce the wounding multiplier of large piercing (pi+), huge piercing (pi++), and impaling (imp) damage to x1. Any major wound (loss of over ⅓ HP from one blow) cripples the extremity. Damage beyond that threshold is lost.");
    }
}
