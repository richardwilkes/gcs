/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.widget.outline.Cell;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.TextCell;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Numbers;

import javax.swing.SwingConstants;

/** Definitions for weapon columns. */
public enum WeaponColumn {
	/** The weapon name/description. */
	DESCRIPTION {
		@Override
		public String toString(Class<? extends WeaponStats> weaponClass) {
			if (weaponClass == MeleeWeaponStats.class) {
				return MELEE;
			}
			if (weaponClass == RangedWeaponStats.class) {
				return RANGED;
			}
			return DESCRIPTION_TITLE;
		}

		@Override
		public String getToolTip() {
			return DESCRIPTION_TOOLTIP;
		}

		@Override
		public Cell getCell(boolean forEditor) {
			return new WeaponDescriptionCell();
		}

		@Override
		public String getDataAsText(WeaponStats weapon) {
			StringBuilder builder = new StringBuilder();
			String notes = weapon.getNotes();

			builder.append(weapon.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
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
			return USAGE_TITLE;
		}

		@Override
		public String getToolTip() {
			return USAGE_TOOLTIP;
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
			return LEVEL_TITLE;
		}

		@Override
		public String getToolTip() {
			return LEVEL_TOOLTIP;
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
			return new Integer(weapon.getSkillLevel());
		}

		@Override
		public String getDataAsText(WeaponStats weapon) {
			int level = weapon.getSkillLevel();

			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return Numbers.format(level);
		}
	},
	/** The weapon accuracy. */
	ACCURACY {
		@Override
		public String toString(Class<? extends WeaponStats> weaponClass) {
			return ACCURACY_TITLE;
		}

		@Override
		public String getToolTip() {
			return ACCURACY_TOOLTIP;
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
			return PARRY_TITLE;
		}

		@Override
		public String getToolTip() {
			return PARRY_TOOLTIP;
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
			return BLOCK_TITLE;
		}

		@Override
		public String getToolTip() {
			return BLOCK_TOOLTIP;
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
			return DAMAGE_TITLE;
		}

		@Override
		public String getToolTip() {
			return DAMAGE_TOOLTIP;
		}

		@Override
		public String getDataAsText(WeaponStats weapon) {
			return weapon.getResolvedDamage();
		}
	},
	/** The weapon reach. */
	REACH {
		@Override
		public String toString(Class<? extends WeaponStats> weaponClass) {
			return REACH_TITLE;
		}

		@Override
		public String getToolTip() {
			return REACH_TOOLTIP;
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
			return RANGE_TITLE;
		}

		@Override
		public String getToolTip() {
			return RANGE_TOOLTIP;
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
			return RATE_OF_FIRE_TITLE;
		}

		@Override
		public String getToolTip() {
			return RATE_OF_FIRE_TOOLTIP;
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
			return SHOTS_TITLE;
		}

		@Override
		public String getToolTip() {
			return SHOTS_TOOLTIP;
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
			return BULK_TITLE;
		}

		@Override
		public String getToolTip() {
			return BULK_TOOLTIP;
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
			return RECOIL_TITLE;
		}

		@Override
		public String getToolTip() {
			return RECOIL_TOOLTIP;
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
			return MIN_ST_TITLE;
		}

		@Override
		public String getToolTip() {
			return MIN_ST_TOOLTIP;
		}

		@Override
		public String getDataAsText(WeaponStats weapon) {
			return weapon.getStrength();
		}
	};

	@Localize("Weapons")
	static String	DESCRIPTION_TITLE;
	@Localize("Melee Weapons")
	static String	MELEE;
	@Localize("Ranged Weapons")
	static String	RANGED;
	@Localize("The name/description of the weapon")
	static String	DESCRIPTION_TOOLTIP;
	@Localize("Usage")
	static String	USAGE_TITLE;
	@Localize("The usage type of the weapon (swung, thrust, thrown, fired, etc.)")
	static String	USAGE_TOOLTIP;
	@Localize("Damage")
	static String	DAMAGE_TITLE;
	@Localize("The damage the weapon inflicts")
	static String	DAMAGE_TOOLTIP;
	@Localize("Reach")
	static String	REACH_TITLE;
	@Localize("The reach of the weapon")
	static String	REACH_TOOLTIP;
	@Localize("Parry")
	static String	PARRY_TITLE;
	@Localize("The Parry value with the weapon")
	static String	PARRY_TOOLTIP;
	@Localize("Block")
	static String	BLOCK_TITLE;
	@Localize("The Block value with the weapon")
	static String	BLOCK_TOOLTIP;
	@Localize("Acc")
	static String	ACCURACY_TITLE;
	@Localize("The accuracy bonus for the weapon")
	static String	ACCURACY_TOOLTIP;
	@Localize("Range")
	static String	RANGE_TITLE;
	@Localize("The range of the weapon")
	static String	RANGE_TOOLTIP;
	@Localize("RoF")
	static String	RATE_OF_FIRE_TITLE;
	@Localize("The rate of fire of the weapon")
	static String	RATE_OF_FIRE_TOOLTIP;
	@Localize("Shots")
	static String	SHOTS_TITLE;
	@Localize("The number of shots the weapon can fire before reloading/recharging")
	static String	SHOTS_TOOLTIP;
	@Localize("Bulk")
	static String	BULK_TITLE;
	@Localize("The modifier to skill due to the bulk of the weapon")
	static String	BULK_TOOLTIP;
	@Localize("Rcl")
	static String	RECOIL_TITLE;
	@Localize("The recoil modifier for the weapon")
	static String	RECOIL_TOOLTIP;
	@Localize("ST")
	static String	MIN_ST_TITLE;
	@Localize("The minimum strength required to use the weapon properly")
	static String	MIN_ST_TOOLTIP;
	@Localize("Lvl")
	static String	LEVEL_TITLE;
	@Localize("The skill level with the weapon")
	static String	LEVEL_TOOLTIP;

	static {
		Localization.initialize();
	}

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
	 * @param forEditor Whether this is for an editor or not.
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
	 * @param outline The {@link Outline} to use.
	 * @param weaponClass The weapon class to use.
	 * @param forEditor Whether this is for an editor or not.
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
