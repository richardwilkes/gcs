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
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.weapon;

import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;
import com.trollworks.ttk.widgets.outline.TextCell;

import javax.swing.SwingConstants;

/** Definitions for weapon columns. */
public enum WeaponColumn {
	/** The weapon name/description. */
	DESCRIPTION {
		@Override
		public String toString(Class<? extends WeaponStats> weaponClass) {
			if (weaponClass == MeleeWeaponStats.class) {
				return MSG_MELEE;
			}
			if (weaponClass == RangedWeaponStats.class) {
				return MSG_RANGED;
			}
			return MSG_DESCRIPTION;
		}

		@Override
		public String getToolTip() {
			return MSG_DESCRIPTION_TOOLTIP;
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
			return MSG_USAGE;
		}

		@Override
		public String getToolTip() {
			return MSG_USAGE_TOOLTIP;
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
			return MSG_LEVEL;
		}

		@Override
		public String getToolTip() {
			return MSG_LEVEL_TOOLTIP;
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
			return MSG_ACCURACY;
		}

		@Override
		public String getToolTip() {
			return MSG_ACCURACY_TOOLTIP;
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
			return MSG_PARRY;
		}

		@Override
		public String getToolTip() {
			return MSG_PARRY_TOOLTIP;
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
			return MSG_BLOCK;
		}

		@Override
		public String getToolTip() {
			return MSG_BLOCK_TOOLTIP;
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
			return MSG_DAMAGE;
		}

		@Override
		public String getToolTip() {
			return MSG_DAMAGE_TOOLTIP;
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
			return MSG_REACH;
		}

		@Override
		public String getToolTip() {
			return MSG_REACH_TOOLTIP;
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
			return MSG_RANGE;
		}

		@Override
		public String getToolTip() {
			return MSG_RANGE_TOOLTIP;
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
			return MSG_ROF;
		}

		@Override
		public String getToolTip() {
			return MSG_ROF_TOOLTIP;
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
			return MSG_SHOTS;
		}

		@Override
		public String getToolTip() {
			return MSG_SHOTS_TOOLTIP;
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
			return MSG_BULK;
		}

		@Override
		public String getToolTip() {
			return MSG_BULK_TOOLTIP;
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
			return MSG_RECOIL;
		}

		@Override
		public String getToolTip() {
			return MSG_RECOIL_TOOLTIP;
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
			return MSG_STRENGTH;
		}

		@Override
		public String getToolTip() {
			return MSG_STRENGTH_TOOLTIP;
		}

		@Override
		public String getDataAsText(WeaponStats weapon) {
			return weapon.getStrength();
		}
	};

	static String	MSG_DESCRIPTION;
	static String	MSG_MELEE;
	static String	MSG_RANGED;
	static String	MSG_DESCRIPTION_TOOLTIP;
	static String	MSG_USAGE;
	static String	MSG_USAGE_TOOLTIP;
	static String	MSG_DAMAGE;
	static String	MSG_DAMAGE_TOOLTIP;
	static String	MSG_REACH;
	static String	MSG_REACH_TOOLTIP;
	static String	MSG_PARRY;
	static String	MSG_PARRY_TOOLTIP;
	static String	MSG_BLOCK;
	static String	MSG_BLOCK_TOOLTIP;
	static String	MSG_ACCURACY;
	static String	MSG_ACCURACY_TOOLTIP;
	static String	MSG_RANGE;
	static String	MSG_RANGE_TOOLTIP;
	static String	MSG_ROF;
	static String	MSG_ROF_TOOLTIP;
	static String	MSG_SHOTS;
	static String	MSG_SHOTS_TOOLTIP;
	static String	MSG_BULK;
	static String	MSG_BULK_TOOLTIP;
	static String	MSG_RECOIL;
	static String	MSG_RECOIL_TOOLTIP;
	static String	MSG_STRENGTH;
	static String	MSG_STRENGTH_TOOLTIP;
	static String	MSG_LEVEL;
	static String	MSG_LEVEL_TOOLTIP;

	static {
		LocalizedMessages.initialize(WeaponColumn.class);
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
	public boolean isValidFor(Class<? extends WeaponStats> weaponClass, boolean forEditor) {
		return true;
	}

	/**
	 * @param weaponClass The weapon class to use.
	 * @return The title of the column.
	 */
	public abstract String toString(Class<? extends WeaponStats> weaponClass);

	@Override
	public String toString() {
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
