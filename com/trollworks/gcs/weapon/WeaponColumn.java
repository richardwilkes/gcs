/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.weapon;

import static com.trollworks.gcs.weapon.WeaponColumn_LS.*;

import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.TextCell;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "DESCRIPTION", msg = "Weapons"),
				@LS(key = "MELEE", msg = "Melee Weapons"),
				@LS(key = "RANGED", msg = "Ranged Weapons"),
				@LS(key = "DESCRIPTION_TOOLTIP", msg = "The name/description of the weapon"),
				@LS(key = "USAGE", msg = "Usage"),
				@LS(key = "USAGE_TOOLTIP", msg = "The usage type of the weapon (swung, thrust, thrown, fired, etc.)"),
				@LS(key = "DAMAGE", msg = "Damage"),
				@LS(key = "DAMAGE_TOOLTIP", msg = "The damage the weapon inflicts"),
				@LS(key = "REACH", msg = "Reach"),
				@LS(key = "REACH_TOOLTIP", msg = "The reach of the weapon"),
				@LS(key = "PARRY", msg = "Parry"),
				@LS(key = "PARRY_TOOLTIP", msg = "The Parry value with the weapon"),
				@LS(key = "BLOCK", msg = "Block"),
				@LS(key = "BLOCK_TOOLTIP", msg = "The Block value with the weapon"),
				@LS(key = "ACCURACY", msg = "Acc"),
				@LS(key = "ACCURACY_TOOLTIP", msg = "The accuracy bonus for the weapon"),
				@LS(key = "RANGE", msg = "Range"),
				@LS(key = "RANGE_TOOLTIP", msg = "The range of the weapon"),
				@LS(key = "RATE_OF_FIRE", msg = "RoF"),
				@LS(key = "RATE_OF_FIRE_TOOLTIP", msg = "The rate of fire of the weapon"),
				@LS(key = "SHOTS", msg = "Shots"),
				@LS(key = "SHOTS_TOOLTIP", msg = "The number of shots the weapon can fire before reloading/recharging"),
				@LS(key = "BULK", msg = "Bulk"),
				@LS(key = "BULK_TOOLTIP", msg = "The modifier to skill due to the bulk of the weapon"),
				@LS(key = "RECOIL", msg = "Rcl"),
				@LS(key = "RECOIL_TOOLTIP", msg = "The recoil modifier for the weapon"),
				@LS(key = "MIN_ST", msg = "ST"),
				@LS(key = "MIN_ST_TOOLTIP", msg = "The minimum strength required to use the weapon properly"),
				@LS(key = "LEVEL", msg = "Lvl"),
				@LS(key = "LEVEL_TOOLTIP", msg = "The skill level with the weapon"),
})
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
			return super.toString(weaponClass);
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
		public String getToolTip() {
			return MIN_ST_TOOLTIP;
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
	public String toString(Class<? extends WeaponStats> weaponClass) {
		return WeaponColumn_LS.toString(this);
	}

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
