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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.model.weapon;

import com.trollworks.gcs.ui.common.CSHeaderCell;
import com.trollworks.gcs.ui.common.CSTextCell;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKTextCell;

/** Definitions for weapon columns. */
public enum CMWeaponColumnID {
	/** The weapon name/description. */
	DESCRIPTION {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			if (weaponClass == CMMeleeWeaponStats.class) {
				return Msgs.MELEE;
			}
			if (weaponClass == CMRangedWeaponStats.class) {
				return Msgs.RANGED;
			}
			return Msgs.DESCRIPTION;
		}

		@Override public String getToolTip() {
			return Msgs.DESCRIPTION_TOOLTIP;
		}

		@Override public TKCell getCell(boolean forEditor) {
			return new CMWeaponDescriptionCell();
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			StringBuilder builder = new StringBuilder();
			String notes = weapon.getNotes();

			builder.append(weapon.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
				builder.append(notes);
			}
			return builder.toString();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return !forEditor;
		}
	},
	/** The weapon usage type. */
	USAGE {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.USAGE;
		}

		@Override public String getToolTip() {
			return Msgs.USAGE_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return weapon.getUsage();
		}
	},
	/** The weapon skill level. */
	LEVEL {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.LEVEL;
		}

		@Override public String getToolTip() {
			return Msgs.LEVEL_TOOLTIP;
		}

		@Override public TKCell getCell(boolean forEditor) {
			if (forEditor) {
				return new CMWeaponTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
			}
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public Object getData(CMWeaponStats weapon) {
			return new Integer(weapon.getSkillLevel());
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			int level = weapon.getSkillLevel();

			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return TKNumberUtils.format(level);
		}
	},
	/** The weapon accuracy. */
	ACCURACY {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.ACCURACY;
		}

		@Override public String getToolTip() {
			return Msgs.ACCURACY_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMRangedWeaponStats) weapon).getAccuracy();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMRangedWeaponStats.class;
		}
	},
	/** The weapon parry. */
	PARRY {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.PARRY;
		}

		@Override public String getToolTip() {
			return Msgs.PARRY_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMMeleeWeaponStats) weapon).getResolvedParry();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMMeleeWeaponStats.class;
		}
	},
	/** The weapon block. */
	BLOCK {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.BLOCK;
		}

		@Override public String getToolTip() {
			return Msgs.BLOCK_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMMeleeWeaponStats) weapon).getResolvedBlock();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMMeleeWeaponStats.class;
		}
	},
	/** The weapon damage. */
	DAMAGE {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.DAMAGE;
		}

		@Override public String getToolTip() {
			return Msgs.DAMAGE_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return weapon.getResolvedDamage();
		}
	},
	/** The weapon reach. */
	REACH {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.REACH;
		}

		@Override public String getToolTip() {
			return Msgs.REACH_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMMeleeWeaponStats) weapon).getReach();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMMeleeWeaponStats.class;
		}
	},
	/** The weapon range. */
	RANGE {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.RANGE;
		}

		@Override public String getToolTip() {
			return Msgs.RANGE_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMRangedWeaponStats) weapon).getResolvedRange();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMRangedWeaponStats.class;
		}
	},
	/** The weapon rate of fire. */
	RATE_OF_FIRE {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.ROF;
		}

		@Override public String getToolTip() {
			return Msgs.ROF_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMRangedWeaponStats) weapon).getRateOfFire();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMRangedWeaponStats.class;
		}
	},
	/** The weapon shots. */
	SHOTS {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.SHOTS;
		}

		@Override public String getToolTip() {
			return Msgs.SHOTS_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMRangedWeaponStats) weapon).getShots();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMRangedWeaponStats.class;
		}
	},
	/** The weapon bulk. */
	BULK {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.BULK;
		}

		@Override public String getToolTip() {
			return Msgs.BULK_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMRangedWeaponStats) weapon).getBulk();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMRangedWeaponStats.class;
		}
	},
	/** The weapon recoil. */
	RECOIL {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.RECOIL;
		}

		@Override public String getToolTip() {
			return Msgs.RECOIL_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return ((CMRangedWeaponStats) weapon).getRecoil();
		}

		@Override public boolean isValidFor(Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
			return weaponClass == CMRangedWeaponStats.class;
		}
	},
	/** The weapon minimum strength. */
	MIN_ST {
		@Override public String toString(Class<? extends CMWeaponStats> weaponClass) {
			return Msgs.STRENGTH;
		}

		@Override public String getToolTip() {
			return Msgs.STRENGTH_TOOLTIP;
		}

		@Override public String getDataAsText(CMWeaponStats weapon) {
			return weapon.getStrength();
		}
	};

	/**
	 * @param weapon The {@link CMWeaponStats} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public Object getData(CMWeaponStats weapon) {
		return getDataAsText(weapon);
	}

	/**
	 * @param weapon The {@link CMWeaponStats} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(CMWeaponStats weapon);

	/** @return The tooltip for the column. */
	public abstract String getToolTip();

	/**
	 * @param forEditor Whether this is for an editor or not.
	 * @return The {@link TKCell} used to display the data.
	 */
	public TKCell getCell(boolean forEditor) {
		if (forEditor) {
			return new CMWeaponTextCell(TKAlignment.LEFT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}
		return new CSTextCell(TKAlignment.LEFT, TKTextCell.COMPARE_AS_TEXT, null, false);
	}

	/**
	 * @param weaponClass The weapon class to check.
	 * @param forEditor Whether this is for an editor or not.
	 * @return Whether this column is valid for the specified weapon class.
	 */
	public boolean isValidFor(@SuppressWarnings("unused") Class<? extends CMWeaponStats> weaponClass, @SuppressWarnings("unused") boolean forEditor) {
		return true;
	}

	/**
	 * @param weaponClass The weapon class to use.
	 * @return The title of the column.
	 */
	public abstract String toString(Class<? extends CMWeaponStats> weaponClass);

	@Override public String toString() {
		return toString(CMWeaponStats.class);
	}

	/**
	 * Adds all relevant {@link TKColumn}s to a {@link TKOutline}.
	 * 
	 * @param outline The {@link TKOutline} to use.
	 * @param weaponClass The weapon class to use.
	 * @param forEditor Whether this is for an editor or not.
	 */
	public static void addColumns(TKOutline outline, Class<? extends CMWeaponStats> weaponClass, boolean forEditor) {
		TKOutlineModel model = outline.getModel();

		for (CMWeaponColumnID one : values()) {
			if (one.isValidFor(weaponClass, forEditor)) {
				TKColumn column = new TKColumn(one.ordinal(), one.toString(weaponClass), one.getToolTip(), one.getCell(forEditor));

				if (!forEditor) {
					column.setHeaderCell(new CSHeaderCell(true));
				}
				model.addColumn(column);
			}
		}
	}
}
