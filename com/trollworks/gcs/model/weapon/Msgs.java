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

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CMWeaponColumnID}. */
	public static String	DESCRIPTION;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	MELEE;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	RANGED;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	DESCRIPTION_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	USAGE;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	USAGE_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	DAMAGE;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	DAMAGE_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	REACH;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	REACH_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	PARRY;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	PARRY_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	BLOCK;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	BLOCK_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	ACCURACY;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	ACCURACY_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	RANGE;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	RANGE_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	ROF;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	ROF_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	SHOTS;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	SHOTS_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	BULK;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	BULK_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	RECOIL;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	RECOIL_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	STRENGTH;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	STRENGTH_TOOLTIP;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	LEVEL;
	/** Used by {@link CMWeaponColumnID}. */
	public static String	LEVEL_TOOLTIP;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
