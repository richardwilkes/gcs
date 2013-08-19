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

package com.trollworks.gcs.ui.spell;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSSpellColumnID}. */
	public static String	SPELLS;
	/** Used by {@link CSSpellColumnID}. */
	public static String	SPELLS_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	CLASS;
	/** Used by {@link CSSpellColumnID}. */
	public static String	CLASS_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	MANA_COST;
	/** Used by {@link CSSpellColumnID}. */
	public static String	MANA_COST_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	TIME;
	/** Used by {@link CSSpellColumnID}. */
	public static String	TIME_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	POINTS;
	/** Used by {@link CSSpellColumnID}. */
	public static String	POINTS_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	LEVEL;
	/** Used by {@link CSSpellColumnID}. */
	public static String	LEVEL_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	RELATIVE_LEVEL;
	/** Used by {@link CSSpellColumnID}. */
	public static String	RELATIVE_LEVEL_TOOLTIP;
	/** Used by {@link CSSpellColumnID}. */
	public static String	REFERENCE;
	/** Used by {@link CSSpellColumnID}. */
	public static String	REFERENCE_TOOLTIP;

	/** Used by {@link CSSpellEditor}. */
	public static String	NAME;
	/** Used by {@link CSSpellEditor}. */
	public static String	NAME_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSSpellEditor}. */
	public static String	TECH_LEVEL;
	/** Used by {@link CSSpellEditor}. */
	public static String	TECH_LEVEL_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	TECH_LEVEL_REQUIRED;
	/** Used by {@link CSSpellEditor}. */
	public static String	TECH_LEVEL_REQUIRED_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	COLLEGE;
	/** Used by {@link CSSpellEditor}. */
	public static String	COLLEGE_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	CLASS_ONLY_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	CLASS_CANNOT_BE_EMPTY;
	/** Used by {@link CSSpellEditor}. */
	public static String	CASTING_COST;
	/** Used by {@link CSSpellEditor}. */
	public static String	CASTING_COST_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	CASTING_COST_CANNOT_BE_EMPTY;
	/** Used by {@link CSSpellEditor}. */
	public static String	MAINTENANCE_COST;
	/** Used by {@link CSSpellEditor}. */
	public static String	MAINTENANCE_COST_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	CASTING_TIME;
	/** Used by {@link CSSpellEditor}. */
	public static String	CASTING_TIME_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	CASTING_TIME_CANNOT_BE_EMPTY;
	/** Used by {@link CSSpellEditor}. */
	public static String	DURATION;
	/** Used by {@link CSSpellEditor}. */
	public static String	DURATION_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	DURATION_CANNOT_BE_EMPTY;
	/** Used by {@link CSSpellEditor}. */
	public static String	NOTES;
	/** Used by {@link CSSpellEditor}. */
	public static String	NOTES_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	EDITOR_POINTS;
	/** Used by {@link CSSpellEditor}. */
	public static String	EDITOR_POINTS_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	EDITOR_LEVEL;
	/** Used by {@link CSSpellEditor}. */
	public static String	EDITOR_LEVEL_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	DIFFICULTY;
	/** Used by {@link CSSpellEditor}. */
	public static String	DIFFICULTY_TOOLTIP;
	/** Used by {@link CSSpellEditor}. */
	public static String	EDITOR_REFERENCE;

	/** Used by {@link CSSpellListWindow}. */
	public static String	UNTITLED;

	/** Used by {@link CSSpellOutline}. */
	public static String	INCREMENT;
	/** Used by {@link CSSpellOutline}. */
	public static String	DECREMENT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
