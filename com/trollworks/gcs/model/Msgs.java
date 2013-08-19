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

package com.trollworks.gcs.model;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CMCharacter}. */
	public static String	HAIR_FORMAT;
	/** Used by {@link CMCharacter}. */
	public static String	BROWN;
	/** Used by {@link CMCharacter}. */
	public static String	BLACK;
	/** Used by {@link CMCharacter}. */
	public static String	BLOND;
	/** Used by {@link CMCharacter}. */
	public static String	REDHEAD;
	/** Used by {@link CMCharacter}. */
	public static String	BALD;
	/** Used by {@link CMCharacter}. */
	public static String	STRAIGHT;
	/** Used by {@link CMCharacter}. */
	public static String	CURLY;
	/** Used by {@link CMCharacter}. */
	public static String	WAVY;
	/** Used by {@link CMCharacter}. */
	public static String	SHORT;
	/** Used by {@link CMCharacter}. */
	public static String	MEDIUM;
	/** Used by {@link CMCharacter}. */
	public static String	LONG;
	/** Used by {@link CMCharacter}. */
	public static String	BLUE;
	/** Used by {@link CMCharacter}. */
	public static String	GREEN;
	/** Used by {@link CMCharacter}. */
	public static String	GREY;
	/** Used by {@link CMCharacter}. */
	public static String	VIOLET;
	/** Used by {@link CMCharacter}. */
	public static String	FRECKLED;
	/** Used by {@link CMCharacter}. */
	public static String	TAN;
	/** Used by {@link CMCharacter}. */
	public static String	LIGHT_TAN;
	/** Used by {@link CMCharacter}. */
	public static String	DARK_TAN;
	/** Used by {@link CMCharacter}. */
	public static String	LIGHT_BROWN;
	/** Used by {@link CMCharacter}. */
	public static String	DARK_BROWN;
	/** Used by {@link CMCharacter}. */
	public static String	PALE;
	/** Used by {@link CMCharacter}. */
	public static String	RIGHT;
	/** Used by {@link CMCharacter}. */
	public static String	LEFT;
	/** Used by {@link CMCharacter}. */
	public static String	MALE;
	/** Used by {@link CMCharacter}. */
	public static String	FEMALE;
	/** Used by {@link CMCharacter}. */
	public static String	DEFAULT_RACE;
	/** Used by {@link CMCharacter}. */
	public static String	LAST_MODIFIED;
	/** Used by {@link CMCharacter}. */
	public static String	CREATED_ON_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	STRENGTH_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	DEXTERITY_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	INTELLIGENCE_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	HEALTH_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	BASIC_SPEED_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	BASIC_MOVE_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	PERCEPTION_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	WILL_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	NAME_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	TITLE_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	AGE_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	BIRTHDAY_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	EYE_COLOR_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	HAIR_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	SKIN_COLOR_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	HANDEDNESS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	HEIGHT_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	WEIGHT_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	GENDER_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	RACE_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	RELIGION_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	PLAYER_NAME_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	CAMPAIGN_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	SIZE_MODIFIER_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	TECH_LEVEL_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	EARNED_POINTS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	HIT_POINTS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	CURRENT_HIT_POINTS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	FATIGUE_POINTS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	CURRENT_FATIGUE_POINTS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	PORTRAIT_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	INCLUDE_PUNCH_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	INCLUDE_KICK_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	INCLUDE_BOOTS_UNDO;
	/** Used by {@link CMCharacter}. */
	public static String	PORTRAIT_COMMENT;
	/** Used by {@link CMCharacter}. */
	public static String	PORTRAIT_WRITE_ERROR;

	/** Used by {@link CMCharacter} and {@link CMTemplate}. */
	public static String	NOTES_UNDO;

	/** Used by {@link CMRowUndo}. */
	public static String	UNDO_FORMAT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
