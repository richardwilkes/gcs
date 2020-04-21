/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** Hit location tables. */
public class HitLocationTable implements Comparable<HitLocationTable> {
    public static final String                        KEY_HUMANOID         = "humanoid";
    public static final String                        KEY_QUADRUPED        = "quadruped";
    public static final String                        KEY_WINGED_QUADRUPED = "winged_quadruped";
    public static final String                        KEY_HEXAPOD          = "hexapod";
    public static final String                        KEY_WINGED_HEXAPOD   = "winged_hexapod";
    public static final String                        KEY_CENTAUR          = "centaur";
    public static final String                        KEY_AVIAN            = "avian";
    public static final String                        KEY_VERMIFORM        = "vermiform";
    public static final String                        KEY_WINGED_VERMIFORM = "winged_vermiform";
    public static final String                        KEY_SNAKEMEN         = "snakemen";
    public static final String                        KEY_OCTOPOD          = "octopod";
    public static final String                        KEY_SQUID            = "squid";
    public static final String                        KEY_CANCROID         = "cancroid";
    public static final String                        KEY_SCORPION         = "scorpion";
    public static final String                        KEY_ICHTHYOID        = "ichthyoid";
    public static final String                        KEY_ARACHNOID        = "arachnoid";
    public static final HitLocationTable              HUMANOID;
    public static final HitLocationTable              QUADRUPED;
    public static final HitLocationTable              WINGED_QUADRUPED;
    public static final HitLocationTable              HEXAPOD;
    public static final HitLocationTable              WINGED_HEXAPOD;
    public static final HitLocationTable              CENTAUR;
    public static final HitLocationTable              AVIAN;
    public static final HitLocationTable              VERMIFORM;
    public static final HitLocationTable              WINGED_VERMIFORM;
    public static final HitLocationTable              SNAKEMEN;
    public static final HitLocationTable              OCTOPOD;
    public static final HitLocationTable              SQUID;
    public static final HitLocationTable              CANCROID;
    public static final HitLocationTable              SCORPION;
    public static final HitLocationTable              ICHTHYOID;
    public static final HitLocationTable              ARACHNOID;
    public static final HitLocationTable[]            ALL;
    public static final Map<String, HitLocationTable> MAP                  = new HashMap<>();

    static {
        List<HitLocationTableEntry> entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Right Leg"), 6, 7));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Right Arm"), 8, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 10));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 11, 11));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Left Arm"), 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Left Leg"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.HAND, 15, 15));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 16, 16));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        HUMANOID = new HitLocationTable(KEY_HUMANOID, I18n.Text("Humanoid"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Foreleg"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Hindleg"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        QUADRUPED = new HitLocationTable(KEY_QUADRUPED, I18n.Text("Quadruped"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Foreleg"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Hindleg"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        WINGED_QUADRUPED = new HitLocationTable(KEY_WINGED_QUADRUPED, I18n.Text("Winged Quadruped"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Foreleg"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 10));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Midleg"), 11, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Hindleg"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Midleg"), 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        HEXAPOD = new HitLocationTable(KEY_HEXAPOD, I18n.Text("Hexapod"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Foreleg"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 10));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Midleg"), 11, 11));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Hindleg"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Midleg"), 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        WINGED_HEXAPOD = new HitLocationTable(KEY_WINGED_HEXAPOD, I18n.Text("Winged Hexapod"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Foreleg"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Hindleg"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.HAND, I18n.Text("Extremity"), 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        CENTAUR = new HitLocationTable(KEY_CENTAUR, I18n.Text("Centaur"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        AVIAN = new HitLocationTable(KEY_AVIAN, I18n.Text("Avian"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        VERMIFORM = new HitLocationTable(KEY_VERMIFORM, I18n.Text("Vermiform"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 14));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 15, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        WINGED_VERMIFORM = new HitLocationTable(KEY_WINGED_VERMIFORM, I18n.Text("Winged Vermiform"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 12));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.HAND, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        SNAKEMEN = new HitLocationTable(KEY_SNAKEMEN, I18n.Text("Snakemen"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE, 1));
        entries.add(new HitLocationTableEntry(HitLocation.BRAIN, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Arm 1-2"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 12));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Arm 3-4"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Arm 5-6"), 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Arm 7-8"), 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        OCTOPOD = new HitLocationTable(KEY_OCTOPOD, I18n.Text("Octopod"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE, 1));
        entries.add(new HitLocationTableEntry(HitLocation.BRAIN, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Arm 1-2"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 12));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, I18n.Text("Extremity"), 13, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        SQUID = new HitLocationTable(KEY_SQUID, I18n.Text("Squid"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, 13, 16));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        CANCROID = new HitLocationTable(KEY_CANCROID, I18n.Text("Cancroid"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, 13, 16));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        SCORPION = new HitLocationTable(KEY_SCORPION, I18n.Text("Scorpion"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE, 1));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FIN, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 7, 12));
        entries.add(new HitLocationTableEntry(HitLocation.FIN, 13, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        ICHTHYOID = new HitLocationTable(KEY_ICHTHYOID, I18n.Text("Ichthyoid"), entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.BRAIN, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Leg 1-2"), 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Leg 3-4"), 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Leg 5-6"), 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, I18n.Text("Leg 7-8"), 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        ARACHNOID = new HitLocationTable(KEY_ARACHNOID, I18n.Text("Arachnoid"), entries);

        ALL = new HitLocationTable[]{HUMANOID, QUADRUPED, WINGED_QUADRUPED, HEXAPOD, WINGED_HEXAPOD, CENTAUR, AVIAN, VERMIFORM, WINGED_VERMIFORM, SNAKEMEN, OCTOPOD, SQUID, CANCROID, SCORPION, ICHTHYOID, ARACHNOID};
        Arrays.sort(ALL);
    }

    private String                      mKey;
    private String                      mName;
    private List<HitLocationTableEntry> mEntries;

    public HitLocationTable(String key, String name, List<HitLocationTableEntry> entries) {
        mKey = key;
        mName = name;
        mEntries = entries;
        MAP.put(mKey, this);
    }

    /** @return The key. */
    public String getKey() {
        return mKey;
    }

    /** @return The name. */
    public String getName() {
        return mName;
    }

    @Override
    public String toString() {
        return mName;
    }

    /** @return The entries. */
    public List<HitLocationTableEntry> getEntries() {
        return mEntries;
    }

    @Override
    public int compareTo(HitLocationTable other) {
        return NumericComparator.caselessCompareStrings(mName, other.mName);
    }
}
