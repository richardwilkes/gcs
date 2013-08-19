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

package com.trollworks.gcs.ui.skills;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSSkillColumnID}. */
	public static String	SKILLS;
	/** Used by {@link CSSkillColumnID}. */
	public static String	SKILLS_TOOLTIP;
	/** Used by {@link CSSkillColumnID}. */
	public static String	LEVEL;
	/** Used by {@link CSSkillColumnID}. */
	public static String	LEVEL_TOOLTIP;
	/** Used by {@link CSSkillColumnID}. */
	public static String	RELATIVE_LEVEL;
	/** Used by {@link CSSkillColumnID}. */
	public static String	RELATIVE_LEVEL_TOOLTIP;
	/** Used by {@link CSSkillColumnID}. */
	public static String	POINTS;
	/** Used by {@link CSSkillColumnID}. */
	public static String	POINTS_TOOLTIP;
	/** Used by {@link CSSkillColumnID}. */
	public static String	DIFFICULTY;
	/** Used by {@link CSSkillColumnID}. */
	public static String	DIFFICULTY_TOOLTIP;
	/** Used by {@link CSSkillColumnID}. */
	public static String	REFERENCE;

	/** Used by {@link CSSkillColumnID} and {@link CSSkillEditor}. */
	public static String	REFERENCE_TOOLTIP;

	/** Used by {@link CSSkillEditor}. */
	public static String	NAME_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSSkillEditor}. */
	public static String	SPECIALIZATION;
	/** Used by {@link CSSkillEditor}. */
	public static String	SPECIALIZATION_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	NOTES_TITLE;
	/** Used by {@link CSSkillEditor}. */
	public static String	TECH_LEVEL;
	/** Used by {@link CSSkillEditor}. */
	public static String	TECH_LEVEL_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	TECH_LEVEL_REQUIRED;
	/** Used by {@link CSSkillEditor}. */
	public static String	TECH_LEVEL_REQUIRED_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	EDITOR_DIFFICULTY_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	EDITOR_DIFFICULTY_POPUP_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	ATTRIBUTE_POPUP_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	EDITOR_POINTS_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	ENC_PENALTY_MULT;
	/** Used by {@link CSSkillEditor}. */
	public static String	ENC_PENALTY_MULT_TOOLTIP;
	/** Used by {@link CSSkillEditor}. */
	public static String	NO_ENC_PENALTY;
	/** Used by {@link CSSkillEditor}. */
	public static String	ONE_ENC_PENALTY;
	/** Used by {@link CSSkillEditor}. */
	public static String	ENC_PENALTY_FORMAT;

	/** Used by {@link CSSkillListWindow}. */
	public static String	UNTITLED;

	/** Used by {@link CSTechniqueEditor}. */
	public static String	SPECIALIZATION_IMPRINT;

	/** Used by {@link CSSkillOutline}. */
	public static String	INCREMENT;
	/** Used by {@link CSSkillOutline}. */
	public static String	DECREMENT;

	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	NAME;
	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	NOTES;
	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	EDITOR_REFERENCE;
	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	EDITOR_POINTS;
	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	EDITOR_LEVEL;
	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	EDITOR_LEVEL_TOOLTIP;
	/** Used by {@link CSSkillEditor} and {@link CSTechniqueEditor}. */
	public static String	EDITOR_DIFFICULTY;

	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_NAME_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_NOTES_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_DIFFICULTY_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_DIFFICULTY_POPUP_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_POINTS_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	TECHNIQUE_REFERENCE_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	DEFAULTS_TO;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	DEFAULTS_TO_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	DEFAULT_NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	DEFAULT_SPECIALIZATION_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	DEFAULT_MODIFIER_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	LIMIT;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	LIMIT_TOOLTIP;
	/** Used by {@link CSTechniqueEditor}. */
	public static String	LIMIT_AMOUNT_TOOLTIP;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
