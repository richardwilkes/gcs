/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.NumericComparator;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** Hit location tables. */
public class HitLocationTable implements Comparable<HitLocationTable> {
    @Localize("Humanoid")
    private static String                             HUMANOID_TITLE;
    @Localize("Quadruped")
    private static String                             QUADRUPED_TITLE;
    @Localize("Winged Quadruped")
    private static String                             WINGED_QUADRUPED_TITLE;
    @Localize("Hexapod")
    private static String                             HEXAPOD_TITLE;
    @Localize("Winged Hexapod")
    private static String                             WINGED_HEXAPOD_TITLE;
    @Localize("Centaur")
    private static String                             CENTAUR_TITLE;
    @Localize("Avian")
    private static String                             AVIAN_TITLE;
    @Localize("Vermiform")
    private static String                             VERMIFORM_TITLE;
    @Localize("Winged Vermiform")
    private static String                             WINGED_VERMIFORM_TITLE;
    @Localize("Snakemen")
    private static String                             SNAKEMEN_TITLE;
    @Localize("Octopod")
    private static String                             OCTOPOD_TITLE;
    @Localize("Squid")
    private static String                             SQUID_TITLE;
    @Localize("Cancroid")
    private static String                             CANCROID_TITLE;
    @Localize("Scorpion")
    private static String                             SCORPION_TITLE;
    @Localize("Ichthyoid")
    private static String                             ICHTHYOID_TITLE;
    @Localize("Arachnoid")
    private static String                             ARACHNOID_TITLE;
    @Localize("Left Leg")
    private static String                             LEFT_LEG_TITLE;
    @Localize("Right Leg")
    private static String                             RIGHT_LEG_TITLE;
    @Localize("Foreleg")
    private static String                             FORELEG_TITLE;
    @Localize("Midleg")
    private static String                             MIDLEG_TITLE;
    @Localize("Hindleg")
    private static String                             HINDLEG_TITLE;
    @Localize("Leg 1-2")
    private static String                             LEG12_TITLE;
    @Localize("Leg 3-4")
    private static String                             LEG34_TITLE;
    @Localize("Leg 5-6")
    private static String                             LEG56_TITLE;
    @Localize("Leg 7-8")
    private static String                             LEG78_TITLE;
    @Localize("Left Arm")
    private static String                             LEFT_ARM_TITLE;
    @Localize("Right Arm")
    private static String                             RIGHT_ARM_TITLE;
    @Localize("Arm 1-2")
    private static String                             ARM12_TITLE;
    @Localize("Arm 3-4")
    private static String                             ARM34_TITLE;
    @Localize("Arm 5-6")
    private static String                             ARM56_TITLE;
    @Localize("Arm 7-8")
    private static String                             ARM78_TITLE;
    @Localize("Extremity")
    private static String                             EXTREMITY_TITLE;

    public static final String                        KEY_HUMANOID         = "humanoid";        			//$NON-NLS-1$
    public static final String                        KEY_QUADRUPED        = "quadruped";       			//$NON-NLS-1$
    public static final String                        KEY_WINGED_QUADRUPED = "winged_quadruped";	//$NON-NLS-1$
    public static final String                        KEY_HEXAPOD          = "hexapod";         			//$NON-NLS-1$
    public static final String                        KEY_WINGED_HEXAPOD   = "winged_hexapod";  		//$NON-NLS-1$
    public static final String                        KEY_CENTAUR          = "centaur";         			//$NON-NLS-1$
    public static final String                        KEY_AVIAN            = "avian";           				//$NON-NLS-1$
    public static final String                        KEY_VERMIFORM        = "vermiform";       			//$NON-NLS-1$
    public static final String                        KEY_WINGED_VERMIFORM = "winged_vermiform";	//$NON-NLS-1$
    public static final String                        KEY_SNAKEMEN         = "snakemen";        			//$NON-NLS-1$
    public static final String                        KEY_OCTOPOD          = "octopod";         			//$NON-NLS-1$
    public static final String                        KEY_SQUID            = "squid";           				//$NON-NLS-1$
    public static final String                        KEY_CANCROID         = "cancroid";        			//$NON-NLS-1$
    public static final String                        KEY_SCORPION         = "scorpion";        			//$NON-NLS-1$
    public static final String                        KEY_ICHTHYOID        = "ichthyoid";       			//$NON-NLS-1$
    public static final String                        KEY_ARACHNOID        = "arachnoid";       			//$NON-NLS-1$

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
        Localization.initialize();

        List<HitLocationTableEntry> entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, RIGHT_LEG_TITLE, 6, 7));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, RIGHT_ARM_TITLE, 8, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 10));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 11, 11));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, LEFT_ARM_TITLE, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, LEFT_LEG_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.HAND, 15, 15));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 16, 16));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        HUMANOID = new HitLocationTable(KEY_HUMANOID, HUMANOID_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, FORELEG_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, HINDLEG_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        QUADRUPED = new HitLocationTable(KEY_QUADRUPED, QUADRUPED_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, FORELEG_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, HINDLEG_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        WINGED_QUADRUPED = new HitLocationTable(KEY_WINGED_QUADRUPED, WINGED_QUADRUPED_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, FORELEG_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 10));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, MIDLEG_TITLE, 11, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, HINDLEG_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, MIDLEG_TITLE, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        HEXAPOD = new HitLocationTable(KEY_HEXAPOD, HEXAPOD_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, FORELEG_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 10));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, MIDLEG_TITLE, 11, 11));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, HINDLEG_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, MIDLEG_TITLE, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.FOOT, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        WINGED_HEXAPOD = new HitLocationTable(KEY_WINGED_HEXAPOD, WINGED_HEXAPOD_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, FORELEG_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, HINDLEG_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.HAND, EXTREMITY_TITLE, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        CENTAUR = new HitLocationTable(KEY_CENTAUR, CENTAUR_TITLE, entries);

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
        AVIAN = new HitLocationTable(KEY_AVIAN, AVIAN_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        VERMIFORM = new HitLocationTable(KEY_VERMIFORM, VERMIFORM_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 14));
        entries.add(new HitLocationTableEntry(HitLocation.WING, 15, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        WINGED_VERMIFORM = new HitLocationTable(KEY_WINGED_VERMIFORM, WINGED_VERMIFORM_TITLE, entries);

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
        SNAKEMEN = new HitLocationTable(KEY_SNAKEMEN, SNAKEMEN_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE, 1));
        entries.add(new HitLocationTableEntry(HitLocation.BRAIN, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, ARM12_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 12));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, ARM34_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, ARM56_TITLE, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, ARM78_TITLE, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        OCTOPOD = new HitLocationTable(KEY_OCTOPOD, OCTOPOD_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE, 1));
        entries.add(new HitLocationTableEntry(HitLocation.BRAIN, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, ARM12_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 12));
        entries.add(new HitLocationTableEntry(HitLocation.ARM, EXTREMITY_TITLE, 13, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        SQUID = new HitLocationTable(KEY_SQUID, SQUID_TITLE, entries);

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
        CANCROID = new HitLocationTable(KEY_CANCROID, CANCROID_TITLE, entries);

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
        SCORPION = new HitLocationTable(KEY_SCORPION, SCORPION_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE, 1));
        entries.add(new HitLocationTableEntry(HitLocation.SKULL, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FIN, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 7, 12));
        entries.add(new HitLocationTableEntry(HitLocation.FIN, 13, 16));
        entries.add(new HitLocationTableEntry(HitLocation.TAIL, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        ICHTHYOID = new HitLocationTable(KEY_ICHTHYOID, ICHTHYOID_TITLE, entries);

        entries = new ArrayList<>();
        entries.add(new HitLocationTableEntry(HitLocation.EYE));
        entries.add(new HitLocationTableEntry(HitLocation.BRAIN, 3, 4));
        entries.add(new HitLocationTableEntry(HitLocation.NECK, 5, 5));
        entries.add(new HitLocationTableEntry(HitLocation.FACE, 6, 6));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, LEG12_TITLE, 7, 8));
        entries.add(new HitLocationTableEntry(HitLocation.TORSO, 9, 11));
        entries.add(new HitLocationTableEntry(HitLocation.GROIN, 12, 12));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, LEG34_TITLE, 13, 14));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, LEG56_TITLE, 15, 16));
        entries.add(new HitLocationTableEntry(HitLocation.LEG, LEG78_TITLE, 17, 18));
        entries.add(new HitLocationTableEntry(HitLocation.VITALS));
        ARACHNOID = new HitLocationTable(KEY_ARACHNOID, ARACHNOID_TITLE, entries);

        ALL = new HitLocationTable[] { HUMANOID, QUADRUPED, WINGED_QUADRUPED, HEXAPOD, WINGED_HEXAPOD, CENTAUR, AVIAN, VERMIFORM, WINGED_VERMIFORM, SNAKEMEN, OCTOPOD, SQUID, CANCROID, SCORPION, ICHTHYOID, ARACHNOID };
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
