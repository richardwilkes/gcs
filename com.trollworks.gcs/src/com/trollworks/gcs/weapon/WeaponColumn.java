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

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.ui.widget.outline.Cell;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListHeaderCell;
import com.trollworks.gcs.ui.widget.outline.ListTextCell;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.TextCell;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import javax.swing.SwingConstants;

/** Definitions for weapon columns. */
public enum WeaponColumn {
    /** The weapon name/description. */
    DESCRIPTION {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            if (weaponClass == MeleeWeaponStats.class) {
                return I18n.Text("Melee Weapons");
            }
            if (weaponClass == RangedWeaponStats.class) {
                return I18n.Text("Ranged Weapons");
            }
            return I18n.Text("Weapons");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The name/description of the weapon");
        }

        @Override
        public Cell getCell(boolean forEditor) {
            return new WeaponDescriptionCell();
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            StringBuilder builder = new StringBuilder();
            String        notes   = weapon.getNotes();

            builder.append(weapon.toString());
            if (!notes.isEmpty()) {
                builder.append(" - ");
                builder.append(notes);
            }
            return builder.toString();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return !forEditor;
        }
    },
    /** The weapon usage type. */
    USAGE {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Usage");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The usage type of the weapon (swung, thrust, thrown, fired, etc.)");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return weapon.getUsage();
        }
    },
    /** The weapon skill level. */
    LEVEL {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Lvl");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The skill level with the weapon");
        }

        @Override
        public Cell getCell(boolean forEditor) {
            if (forEditor) {
                return new TextCell(SwingConstants.RIGHT, false);
            }
            return new ListTextCell(SwingConstants.RIGHT, false);
        }

        @Override
        public Object getData(WeaponStats weapon) {
            return Integer.valueOf(weapon.getSkillLevel());
        }

        @Override
        public String getToolTip(WeaponDisplayRow weapon) {
            return weapon.getSkillLevelToolTip();
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            int level = weapon.getSkillLevel();
            if (level < 0) {
                return "-";
            }
            return Numbers.format(level);
        }
    },
    /** The weapon accuracy. */
    ACCURACY {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Acc");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The accuracy bonus for the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((RangedWeaponStats) weapon).getAccuracy();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == RangedWeaponStats.class;
        }
    },
    /** The weapon parry. */
    PARRY {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Parry");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The Parry value with the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((MeleeWeaponStats) weapon).getResolvedParry();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == MeleeWeaponStats.class;
        }
    },
    /** The weapon block. */
    BLOCK {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Block");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The Block value with the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((MeleeWeaponStats) weapon).getResolvedBlock();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == MeleeWeaponStats.class;
        }
    },
    /** The weapon damage. */
    DAMAGE {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Damage");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The damage the weapon inflicts");
        }

        @Override
        public String getToolTip(WeaponDisplayRow weapon) {
            return weapon.getDamageToolTip();
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return weapon.getDamage().getResolvedDamage();
        }
    },
    /** The weapon reach. */
    REACH {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Reach");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The reach of the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((MeleeWeaponStats) weapon).getReach();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == MeleeWeaponStats.class;
        }
    },
    /** The weapon range. */
    RANGE {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Range");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The range of the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((RangedWeaponStats) weapon).getResolvedRange();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == RangedWeaponStats.class;
        }
    },
    /** The weapon rate of fire. */
    RATE_OF_FIRE {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("RoF");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The rate of fire of the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((RangedWeaponStats) weapon).getRateOfFire();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == RangedWeaponStats.class;
        }
    },
    /** The weapon shots. */
    SHOTS {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Shots");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The number of shots the weapon can fire before reloading/recharging");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((RangedWeaponStats) weapon).getShots();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == RangedWeaponStats.class;
        }
    },
    /** The weapon bulk. */
    BULK {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Bulk");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The modifier to skill due to the bulk of the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((RangedWeaponStats) weapon).getBulk();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == RangedWeaponStats.class;
        }
    },
    /** The weapon recoil. */
    RECOIL {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("Rcl");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The recoil modifier for the weapon");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return ((RangedWeaponStats) weapon).getRecoil();
        }

        @Override
        public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
            return weaponClass == RangedWeaponStats.class;
        }
    },
    /** The weapon minimum strength. */
    MIN_ST {
        @Override
        public String toString(Class<? extends WeaponStats> weaponClass) {
            return I18n.Text("ST");
        }

        @Override
        public String getToolTip() {
            return I18n.Text("The minimum strength required to use the weapon properly");
        }

        @Override
        public String getDataAsText(WeaponStats weapon) {
            return weapon.getStrength();
        }
    };

    /**
     * @param weapon The {@link WeaponStats} to get the data from.
     * @return An object representing the data for this column.
     */
    public Object getData(WeaponStats weapon) {
        return getDataAsText(weapon);
    }

    /**
     * @param weapon The {@link WeaponStats} to get the data from.
     * @return Text representing the data for this column.
     */
    public abstract String getDataAsText(WeaponStats weapon);

    /** @return The tooltip for the column. */
    public abstract String getToolTip();

    /**
     * @param weapon The {@link WeaponDisplayRow} to get the data from.
     * @return The tooltip for a specific row within the column.
     */
    @SuppressWarnings("static-method")
    public String getToolTip(WeaponDisplayRow weapon) {
        return null;
    }

    /**
     * @param forEditor Whether this is for an editor or not.
     * @return The {@link Cell} used to display the data.
     */
    @SuppressWarnings("static-method")
    public Cell getCell(boolean forEditor) {
        if (forEditor) {
            return new TextCell(SwingConstants.LEFT, false);
        }
        return new ListTextCell(SwingConstants.LEFT, false);
    }

    /**
     * @param weaponClass The weapon class to check.
     * @param forEditor   Whether this is for an editor or not.
     * @return Whether this column is valid for the specified weapon class.
     */
    @SuppressWarnings("static-method")
    public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
        return true;
    }

    /**
     * @param weaponClass The weapon class to use.
     * @return The title of the column.
     */
    public abstract String toString(Class<? extends WeaponStats> weaponClass);

    @Override
    public final String toString() {
        return toString(WeaponStats.class);
    }

    /**
     * Adds all relevant {@link Column}s to a {@link Outline}.
     *
     * @param outline     The {@link Outline} to use.
     * @param weaponClass The weapon class to use.
     * @param forEditor   Whether this is for an editor or not.
     */
    public static void addColumns(Outline outline, Class<? extends WeaponStats> weaponClass, boolean forEditor) {
        OutlineModel model = outline.getModel();
        for (WeaponColumn one : values()) {
            if (one.isValidFor(weaponClass, forEditor)) {
                Column column = new Column(one.ordinal(), one.toString(weaponClass), one.getToolTip(), one.getCell(forEditor));
                if (!forEditor) {
                    column.setHeaderCell(new ListHeaderCell(true));
                }
                model.addColumn(column);
            }
        }
    }
}
