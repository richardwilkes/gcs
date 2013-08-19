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
	/** Used by {@link CSAttributeBonus}. */
	public static String	ST;
	/** Used by {@link CSAttributeBonus}. */
	public static String	DX;
	/** Used by {@link CSAttributeBonus}. */
	public static String	IQ;
	/** Used by {@link CSAttributeBonus}. */
	public static String	HT;
	/** Used by {@link CSAttributeBonus}. */
	public static String	WILL;
	/** Used by {@link CSAttributeBonus}. */
	public static String	PERCEPTION;
	/** Used by {@link CSAttributeBonus}. */
	public static String	VISION;
	/** Used by {@link CSAttributeBonus}. */
	public static String	HEARING;
	/** Used by {@link CSAttributeBonus}. */
	public static String	TASTE_SMELL;
	/** Used by {@link CSAttributeBonus}. */
	public static String	TOUCH;
	/** Used by {@link CSAttributeBonus}. */
	public static String	DODGE;
	/** Used by {@link CSAttributeBonus}. */
	public static String	PARRY;
	/** Used by {@link CSAttributeBonus}. */
	public static String	BLOCK;
	/** Used by {@link CSAttributeBonus}. */
	public static String	SPEED;
	/** Used by {@link CSAttributeBonus}. */
	public static String	MOVE;
	/** Used by {@link CSAttributeBonus}. */
	public static String	FP;
	/** Used by {@link CSAttributeBonus}. */
	public static String	HP;
	/** Used by {@link CSAttributeBonus}. */
	public static String	SM;
	/** Used by {@link CSAttributeBonus}. */
	public static String	NO_LIMITATION;
	/** Used by {@link CSAttributeBonus}. */
	public static String	STRIKING_ONLY;
	/** Used by {@link CSAttributeBonus}. */
	public static String	LIFTING_ONLY;

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
	public static String	PER_LEVEL;

	/** Used by {@link CSCostReduction}. */
	public static String	BY;

	/** Used by {@link CSDRBonus}. */
	public static String	SKULL;
	/** Used by {@link CSDRBonus}. */
	public static String	EYES;
	/** Used by {@link CSDRBonus}. */
	public static String	FACE;
	/** Used by {@link CSDRBonus}. */
	public static String	NECK;
	/** Used by {@link CSDRBonus}. */
	public static String	TORSO;
	/** Used by {@link CSDRBonus}. */
	public static String	FULL_BODY;
	/** Used by {@link CSDRBonus}. */
	public static String	FULL_BODY_NO_EYES;
	/** Used by {@link CSDRBonus}. */
	public static String	GROIN;
	/** Used by {@link CSDRBonus}. */
	public static String	ARMS;
	/** Used by {@link CSDRBonus}. */
	public static String	HANDS;
	/** Used by {@link CSDRBonus}. */
	public static String	LEGS;
	/** Used by {@link CSDRBonus}. */
	public static String	FEET;

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

	static {
		TKMessages.initialize(Msgs.class);
	}
}
