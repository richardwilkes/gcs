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

package com.trollworks.gcs.ui.editor.feature;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSBaseFeature}. */
	public static String	ADD_FEATURE_TOOLTIP;
	/** Used by {@link CSBaseFeature}. */
	public static String	REMOVE_FEATURE_TOOLTIP;
	/** Used by {@link CSBaseFeature}. */
	public static String	ATTRIBUTE_BONUS;
	/** Used by {@link CSBaseFeature}. */
	public static String	COST_REDUCTION;
	/** Used by {@link CSBaseFeature}. */
	public static String	DR_BONUS;
	/** Used by {@link CSBaseFeature}. */
	public static String	SKILL_BONUS;
	/** Used by {@link CSBaseFeature}. */
	public static String	SPELL_BONUS;
	/** Used by {@link CSBaseFeature}. */
	public static String	WEAPON_BONUS;
	/** Used by {@link CSBaseFeature}. */
	public static String	PER_LEVEL;
	/** Used by {@link CSBaseFeature}. */
	public static String	PER_DIE;

	/** Used by {@link CSCostReduction}. */
	public static String	BY;

	/** Used by {@link CSFeatures}. */
	public static String	FEATURES;

	/** Used by {@link CSSkillBonus}. */
	public static String	SKILL_NAME;
	/** Used by {@link CSSkillBonus}. */
	public static String	SPECIALIZATION;

	/** Used by {@link CSSpellBonus}. */
	public static String	ALL_COLLEGES;
	/** Used by {@link CSSpellBonus}. */
	public static String	ONE_COLLEGE;
	/** Used by {@link CSSpellBonus}. */
	public static String	SPELL_NAME;

	/** Used by {@link CSWeaponBonus}. */
	public static String	WEAPON_SKILL;
	/** Used by {@link CSWeaponBonus}. */
	public static String	RELATIVE_SKILL_LEVEL;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
