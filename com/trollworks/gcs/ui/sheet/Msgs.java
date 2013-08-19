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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSAttributesPanel}. */
	public static String	ATTRIBUTES;
	/** Used by {@link CSAttributesPanel}. */
	public static String	ST;
	/** Used by {@link CSAttributesPanel}. */
	public static String	ST_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	DX;
	/** Used by {@link CSAttributesPanel}. */
	public static String	DX_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	IQ;
	/** Used by {@link CSAttributesPanel}. */
	public static String	IQ_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	HT;
	/** Used by {@link CSAttributesPanel}. */
	public static String	HT_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	WILL;
	/** Used by {@link CSAttributesPanel}. */
	public static String	WILL_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	PERCEPTION;
	/** Used by {@link CSAttributesPanel}. */
	public static String	PERCEPTION_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	VISION;
	/** Used by {@link CSAttributesPanel}. */
	public static String	HEARING;
	/** Used by {@link CSAttributesPanel}. */
	public static String	TOUCH;
	/** Used by {@link CSAttributesPanel}. */
	public static String	TASTE_SMELL;
	/** Used by {@link CSAttributesPanel}. */
	public static String	BASIC_SPEED;
	/** Used by {@link CSAttributesPanel}. */
	public static String	BASIC_SPEED_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	BASIC_MOVE;
	/** Used by {@link CSAttributesPanel}. */
	public static String	BASIC_MOVE_TOOLTIP;
	/** Used by {@link CSAttributesPanel}. */
	public static String	THRUST;
	/** Used by {@link CSAttributesPanel}. */
	public static String	SWING;

	/** Used by {@link CSDescriptionPanel}. */
	public static String	DESCRIPTION;
	/** Used by {@link CSDescriptionPanel}. */
	public static String	RACE;
	/** Used by {@link CSDescriptionPanel}. */
	public static String	SIZE_MODIFIER;
	/** Used by {@link CSDescriptionPanel}. */
	public static String	SIZE_MODIFIER_TOOLTIP;
	/** Used by {@link CSDescriptionPanel}. */
	public static String	TECH_LEVEL;
	/** Used by {@link CSDescriptionPanel}. */
	public static String	TECH_LEVEL_TOOLTIP;

	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	AGE;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	GENDER;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	BIRTHDAY;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	HEIGHT;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	WEIGHT;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	HAIR;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	HAIR_TOOLTIP;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	EYE_COLOR;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	EYE_COLOR_TOOLTIP;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	SKIN_COLOR;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	SKIN_COLOR_TOOLTIP;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	HANDEDNESS;
	/** Used by {@link CSDescriptionPanel} and {@link CSDescriptionRandomizer}. */
	public static String	HANDEDNESS_TOOLTIP;

	/** Used by {@link CSDescriptionRandomizer}. */
	public static String	RANDOMIZER;
	/** Used by {@link CSDescriptionRandomizer}. */
	public static String	APPLY;
	/** Used by {@link CSDescriptionRandomizer}. */
	public static String	RANDOMIZE;
	/** Used by {@link CSDescriptionRandomizer}. */
	public static String	UNDO_RANDOMIZE;

	/** Used by {@link CSEncumbrancePanel}. */
	public static String	ENCUMBRANCE_MOVE_DODGE;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	ENCUMBRANCE_LEVEL;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	MAX_CARRY;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	MOVE;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	DODGE;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	ENCUMBRANCE_TOOLTIP;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	MAX_CARRY_TOOLTIP;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	MOVE_TOOLTIP;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	DODGE_TOOLTIP;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	NONE;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	LIGHT;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	MEDIUM;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	HEAVY;
	/** Used by {@link CSEncumbrancePanel}. */
	public static String	EXTRA_HEAVY;

	/** Used by {@link CSEncumbrancePanel} and {@link CSSheet}. */
	public static String	ENCUMBRANCE_FORMAT;
	/** Used by {@link CSEncumbrancePanel} and {@link CSSheet}. */
	public static String	CURRENT_ENCUMBRANCE_FORMAT;

	/** Used by {@link CSHitLocationPanel}. */
	public static String	HIT_LOCATION;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	ROLL;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	ROLL_TOOLTIP;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	LOCATION;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	PENALTY;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	PENALTY_TITLE_TOOLTIP;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	PENALTY_TOOLTIP;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	DR;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	DR_TOOLTIP;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	EYE;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	SKULL;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	FACE;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	RIGHT_LEG;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	RIGHT_ARM;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	TORSO;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	GROIN;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	LEFT_ARM;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	LEFT_LEG;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	HAND;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	FOOT;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	NECK;
	/** Used by {@link CSHitLocationPanel}. */
	public static String	VITALS;

	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_HP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_CURRENT;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_CURRENT_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_REELING;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_REELING_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_UNCONSCIOUS_CHECKS;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_UNCONSCIOUS_CHECKS_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_1;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_1_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_2;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_2_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_3;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_3_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_4;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEATH_CHECK_4_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEAD;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	HP_DEAD_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_CURRENT;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_CURRENT_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_TIRED;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_TIRED_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_UNCONSCIOUS_CHECKS;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_UNCONSCIOUS_CHECKS_TOOLTIP;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_UNCONSCIOUS;
	/** Used by {@link CSHitPointsPanel}. */
	public static String	FP_UNCONSCIOUS_TOOLTIP;

	/** Used by {@link CSIdentityPanel}. */
	public static String	IDENTITY;
	/** Used by {@link CSIdentityPanel}. */
	public static String	NAME;
	/** Used by {@link CSIdentityPanel}. */
	public static String	TITLE;
	/** Used by {@link CSIdentityPanel}. */
	public static String	RELIGION;

	/** Used by {@link CSLiftPanel}. */
	public static String	LIFT_MOVE;
	/** Used by {@link CSLiftPanel}. */
	public static String	BASIC_LIFT;
	/** Used by {@link CSLiftPanel}. */
	public static String	BASIC_LIFT_TOOLTIP;
	/** Used by {@link CSLiftPanel}. */
	public static String	ONE_HANDED_LIFT;
	/** Used by {@link CSLiftPanel}. */
	public static String	ONE_HANDED_LIFT_TOOLTIP;
	/** Used by {@link CSLiftPanel}. */
	public static String	TWO_HANDED_LIFT;
	/** Used by {@link CSLiftPanel}. */
	public static String	TWO_HANDED_LIFT_TOOLTIP;
	/** Used by {@link CSLiftPanel}. */
	public static String	SHOVE_KNOCK_OVER;
	/** Used by {@link CSLiftPanel}. */
	public static String	SHOVE_KNOCK_OVER_TOOLTIP;
	/** Used by {@link CSLiftPanel}. */
	public static String	RUNNING_SHOVE;
	/** Used by {@link CSLiftPanel}. */
	public static String	RUNNING_SHOVE_TOOLTIP;
	/** Used by {@link CSLiftPanel}. */
	public static String	CARRY_ON_BACK;
	/** Used by {@link CSLiftPanel}. */
	public static String	CARRY_ON_BACK_TOOLTIP;
	/** Used by {@link CSLiftPanel}. */
	public static String	SHIFT_SLIGHTLY;
	/** Used by {@link CSLiftPanel}. */
	public static String	SHIFT_SLIGHTLY_TOOLTIP;

	/** Used by {@link CSNotesPanel}. */
	public static String	NOTES_CONTINUED;
	/** Used by {@link CSNotesPanel}. */
	public static String	NOTES_TOOLTIP;

	/** Used by {@link CSPlayerInfoPanel}. */
	public static String	PLAYER_INFO;
	/** Used by {@link CSPlayerInfoPanel}. */
	public static String	PLAYER_NAME;
	/** Used by {@link CSPlayerInfoPanel}. */
	public static String	CAMPAIGN;
	/** Used by {@link CSPlayerInfoPanel}. */
	public static String	CREATED_ON;

	/** Used by {@link CSPointsPanel}. */
	public static String	POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	ATTRIBUTE_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	ATTRIBUTE_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	ADVANTAGE_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	ADVANTAGE_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	DISADVANTAGE_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	DISADVANTAGE_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	QUIRK_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	QUIRK_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	SKILL_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	SKILL_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	SPELL_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	SPELL_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	RACE_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	RACE_POINTS_TOOLTIP;
	/** Used by {@link CSPointsPanel}. */
	public static String	EARNED_POINTS;
	/** Used by {@link CSPointsPanel}. */
	public static String	EARNED_POINTS_TOOLTIP;

	/** Used by {@link CSPortraitPanel}. */
	public static String	IMAGE_FILES;
	/** Used by {@link CSPortraitPanel}. */
	public static String	PORTRAIT;
	/** Used by {@link CSPortraitPanel}. */
	public static String	PORTRAIT_TOOLTIP;
	/** Used by {@link CSPortraitPanel}. */
	public static String	BAD_IMAGE;

	/** Used by {@link CSPrerequisitesThread}. */
	public static String	BULLET_PREFIX;

	/** Used by {@link CSSheetOpener}. */
	public static String	CHARACTER_SHEETS;

	/** Used by {@link CSSheet}. */
	public static String	PAGE_NUMBER;
	/** Used by {@link CSSheet}. */
	public static String	ADVERTISEMENT;
	/** Used by {@link CSSheet}. */
	public static String	MELEE_WEAPONS;
	/** Used by {@link CSSheet}. */
	public static String	RANGED_WEAPONS;
	/** Used by {@link CSSheet}. */
	public static String	ADVANTAGES;
	/** Used by {@link CSSheet}. */
	public static String	SKILLS;
	/** Used by {@link CSSheet}. */
	public static String	SPELLS;
	/** Used by {@link CSSheet}. */
	public static String	CARRIED_EQUIPMENT;
	/** Used by {@link CSSheet}. */
	public static String	OTHER_EQUIPMENT;
	/** Used by {@link CSSheet}. */
	public static String	CONTINUED;
	/** Used by {@link CSSheet}. */
	public static String	NATURAL;
	/** Used by {@link CSSheet}. */
	public static String	PUNCH;
	/** Used by {@link CSSheet}. */
	public static String	KICK;
	/** Used by {@link CSSheet}. */
	public static String	BOOTS;
	/** Used by {@link CSSheet}. */
	public static String	UNIDENTIFIED_KEY;

	/** Used by {@link CSNotesPanel} and {@link CSSheet}. */
	public static String	NOTES;

	/** Used by {@link CSSheetWindow}. */
	public static String	PNG_DESCRIPTION;
	/** Used by {@link CSSheetWindow}. */
	public static String	PDF_DESCRIPTION;
	/** Used by {@link CSSheetWindow}. */
	public static String	HTML_DESCRIPTION;
	/** Used by {@link CSSheetWindow}. */
	public static String	SAVE_AS_PNG_ERROR;
	/** Used by {@link CSSheetWindow}. */
	public static String	SAVE_AS_PDF_ERROR;
	/** Used by {@link CSSheetWindow}. */
	public static String	SAVE_AS_HTML_ERROR;
	/** Used by {@link CSSheetWindow}. */
	public static String	SAVE_ERROR;
	/** Used by {@link CSSheetWindow}. */
	public static String	UNTITLED_SHEET;
	/** Used by {@link CSSheetWindow}. */
	public static String	ADD_ROWS;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
