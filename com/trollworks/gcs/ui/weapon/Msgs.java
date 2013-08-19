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

package com.trollworks.gcs.ui.weapon;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSMeleeWeaponEditor}. */
	public static String	MELEE_WEAPON;
	/** Used by {@link CSMeleeWeaponEditor}. */
	public static String	REACH;
	/** Used by {@link CSMeleeWeaponEditor}. */
	public static String	PARRY;
	/** Used by {@link CSMeleeWeaponEditor}. */
	public static String	BLOCK;

	/** Used by {@link CSMeleeWeaponEditor} and {@link CSRangedWeaponEditor}. */
	public static String	USAGE;
	/** Used by {@link CSMeleeWeaponEditor} and {@link CSRangedWeaponEditor}. */
	public static String	DAMAGE;
	/** Used by {@link CSMeleeWeaponEditor} and {@link CSRangedWeaponEditor}. */
	public static String	MINIMUM_STRENGTH;

	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	RANGED_WEAPON;
	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	ACCURACY;
	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	RANGE;
	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	RATE_OF_FIRE;
	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	SHOTS;
	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	BULK;
	/** Used by {@link CSRangedWeaponEditor}. */
	public static String	RECOIL;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
