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
import com.trollworks.gcs.utility.I18n;

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

    public static final List<HitLocationTable> createStdTables() {
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

    public static final HitLocationTable createArachnoidTable() {
        HitLocationTable table = new HitLocationTable("arachnoid", I18n.Text("Arachnoid"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("brain", I18n.Text("Brain"), 2, -7, 1, getBrainDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg 1-2"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg 3-4"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg 5-6"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg 7-8"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createAvianTable() {
        HitLocationTable table = new HitLocationTable("avian", I18n.Text("Avian"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
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
        return table;
    }

    public static final HitLocationTable createCancroidTable() {
        HitLocationTable table = new HitLocationTable("cancroid", I18n.Text("Cancroid"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), 4, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createCentaurTable() {
        HitLocationTable table = new HitLocationTable("centaur", I18n.Text("Centaur"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("hand", I18n.Text("Extremity"), 2, -4, 0, getHandDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createHexapodTable() {
        HitLocationTable table = new HitLocationTable("hexapod", I18n.Text("Hexapod"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("leg", I18n.Text("Midleg"), 1, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Midleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createWingedHexapodTable() {
        HitLocationTable table = new HitLocationTable("hexapod.winged", I18n.Text("Hexapod, Winged"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("leg", I18n.Text("Midleg"), 1, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 1, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Midleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createHumanoidTable() {
        HitLocationTable table = new HitLocationTable("humanoid", I18n.Text("Humanoid"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Right Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Right Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Left Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Left Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("hand", I18n.Text("Hand"), 1, -4, 0, getHandDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 1, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 2, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createIchthyoidTable() {
        HitLocationTable table = new HitLocationTable("ichthyoid", I18n.Text("Ichthyoid"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -8, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("fin", I18n.Text("Fin"), 1, -4, 0, getFinDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 6, 0, 0, ""));
        table.addLocation(new HitLocation("fin", I18n.Text("Fin"), 4, -4, 0, getFinDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createOctopodTable() {
        HitLocationTable table = new HitLocationTable("octopod", I18n.Text("Octopod"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -8, 0, getEyeDescription()));
        table.addLocation(new HitLocation("brain", I18n.Text("Brain"), 2, -7, 1, getBrainDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm 1-2"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm 3-4"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm 5-6"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm 7-8"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createQuadrupedTable() {
        HitLocationTable table = new HitLocationTable("quadruped", I18n.Text("Quadruped"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, getGroinDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createWingedQuadrupedTable() {
        HitLocationTable table = new HitLocationTable("quadruped.winged", I18n.Text("Quadruped, Winged"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Foreleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 1, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Hindleg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 2, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }


    public static final HitLocationTable createScorpionTable() {
        HitLocationTable table = new HitLocationTable("scorpion", I18n.Text("scorpion"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 3, 0, 0, ""));
        table.addLocation(new HitLocation("tail", I18n.Text("Tail"), 1, -3, 0, getTailDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), 4, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 2, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createSnakemenTable() {
        HitLocationTable table = new HitLocationTable("snakemen", I18n.Text("Snakemen"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("hand", I18n.Text("Hand"), 2, -4, 0, getHandDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createSquidTable() {
        HitLocationTable table = new HitLocationTable("squid", I18n.Text("Squid"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -8, 0, getEyeDescription()));
        table.addLocation(new HitLocation("brain", I18n.Text("Brain"), 2, -7, 1, getBrainDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 1, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm 1-2"), 2, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 4, 0, 0, ""));
        table.addLocation(new HitLocation("hand", I18n.Text("Extremity"), 4, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createVeriformTable() {
        HitLocationTable table = new HitLocationTable("vermiform", I18n.Text("Vermiform"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 3, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 10, 0, 0, ""));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
        return table;
    }

    public static final HitLocationTable createWingedVeriformTable() {
        HitLocationTable table = new HitLocationTable("vermiform.winged", I18n.Text("Vermiform, Winged"), new Dice(3));
        table.addLocation(new HitLocation("eyes", I18n.Text("Eyes"), 0, -9, 0, getEyeDescription()));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, getSkullDescription()));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, getFaceDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 3, -5, 0, getNeckDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 6, 0, 0, ""));
        table.addLocation(new HitLocation("wing", I18n.Text("Wing"), 4, -2, 0, getWingDescription()));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, getVitalsDescription()));
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
