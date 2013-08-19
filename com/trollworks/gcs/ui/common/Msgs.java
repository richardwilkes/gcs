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

package com.trollworks.gcs.ui.common;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSAboutPanel}. */
	public static String	TESTERS;
	/** Used by {@link CSAboutPanel}. */
	public static String	TITLE;
	/** Used by {@link CSAboutPanel}. */
	public static String	VERSION;
	/** Used by {@link CSAboutPanel}. */
	public static String	THANKS;
	/** Used by {@link CSAboutPanel}. */
	public static String	GURPS_LICENSE;
	/** Used by {@link CSAboutPanel}. */
	public static String	ITEXT_LICENSE;

	/** Used by {@link CSFileOpener}. */
	public static String	FILES_DESCRIPTION;

	/** Used by {@link CSListOpener}. */
	public static String	ADVANTAGES_DESCRIPTION;
	/** Used by {@link CSListOpener}. */
	public static String	EQUIPMENT_DESCRIPTION;
	/** Used by {@link CSListOpener}. */
	public static String	SKILLS_DESCRIPTION;
	/** Used by {@link CSListOpener}. */
	public static String	SPELLS_DESCRIPTION;
	/** Used by {@link CSListOpener}. */
	public static String	CHOICE_QUERY;
	/** Used by {@link CSListOpener}. */
	public static String	ADVANTAGE_CHOICE;
	/** Used by {@link CSListOpener}. */
	public static String	SKILL_CHOICE;
	/** Used by {@link CSListOpener}. */
	public static String	SPELL_CHOICE;
	/** Used by {@link CSListOpener}. */
	public static String	EQUIPMENT_CHOICE;

	/** Used by {@link CSListWindow}. */
	public static String	TOGGLE_ROWS_OPEN_TOOLTIP;
	/** Used by {@link CSListWindow}. */
	public static String	SIZE_COLUMNS_TO_FIT_TOOLTIP;
	/** Used by {@link CSListWindow}. */
	public static String	TOGGLE_EDIT_MODE_TOOLTIP;
	/** Used by {@link CSListWindow}. */
	public static String	SAVE_ERROR;

	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_CHARACTER_SHEET;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_CHARACTER_TEMPLATE;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_LIST;
	/** Used by {@link CSMenuKeys}. */
	public static String	OPEN;
	/** Used by {@link CSMenuKeys}. */
	public static String	CLOSE;
	/** Used by {@link CSMenuKeys}. */
	public static String	SAVE;
	/** Used by {@link CSMenuKeys}. */
	public static String	SAVE_AS;
	/** Used by {@link CSMenuKeys}. */
	public static String	PAGE_SETUP;
	/** Used by {@link CSMenuKeys}. */
	public static String	PRINT;
	/** Used by {@link CSMenuKeys}. */
	public static String	UNDO;
	/** Used by {@link CSMenuKeys}. */
	public static String	REDO;
	/** Used by {@link CSMenuKeys}. */
	public static String	CUT;
	/** Used by {@link CSMenuKeys}. */
	public static String	COPY;
	/** Used by {@link CSMenuKeys}. */
	public static String	PASTE;
	/** Used by {@link CSMenuKeys}. */
	public static String	DUPLICATE;
	/** Used by {@link CSMenuKeys}. */
	public static String	CLEAR;
	/** Used by {@link CSMenuKeys}. */
	public static String	SELECT_ALL;
	/** Used by {@link CSMenuKeys}. */
	public static String	INCREMENT;
	/** Used by {@link CSMenuKeys}. */
	public static String	DECREMENT;
	/** Used by {@link CSMenuKeys}. */
	public static String	TOGGLE_EQUIPPED;
	/** Used by {@link CSMenuKeys}. */
	public static String	JUMP_TO_FIND;
	/** Used by {@link CSMenuKeys}. */
	public static String	RANDOMIZE_DESCRIPTION;
	/** Used by {@link CSMenuKeys}. */
	public static String	RANDOMIZE_FEMALE_NAME;
	/** Used by {@link CSMenuKeys}. */
	public static String	RANDOMIZE_MALE_NAME;
	/** Used by {@link CSMenuKeys}. */
	public static String	ADD_NATURAL_PUNCH;
	/** Used by {@link CSMenuKeys}. */
	public static String	ADD_NATURAL_KICK;
	/** Used by {@link CSMenuKeys}. */
	public static String	ADD_NATURAL_KICK_WITH_BOOTS;
	/** Used by {@link CSMenuKeys}. */
	public static String	RESET_COLUMNS;
	/** Used by {@link CSMenuKeys}. */
	public static String	RESET_CONFIRMATION_DIALOGS;
	/** Used by {@link CSMenuKeys}. */
	public static String	PREFERENCES;
	/** Used by {@link CSMenuKeys}. */
	public static String	OPEN_EDITOR;
	/** Used by {@link CSMenuKeys}. */
	public static String	COPY_TO_SHEET;
	/** Used by {@link CSMenuKeys}. */
	public static String	COPY_TO_TEMPLATE;
	/** Used by {@link CSMenuKeys}. */
	public static String	APPLY_TEMPLATE_TO_SHEET;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_ADVANTAGE;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_ADVANTAGE_CONTAINER;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_SKILL;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_SKILL_CONTAINER;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_TECHNIQUE;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_SPELL;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_SPELL_CONTAINER;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_CARRIED_EQUIPMENT;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_CARRIED_EQUIPMENT_CONTAINER;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_EQUIPMENT;
	/** Used by {@link CSMenuKeys}. */
	public static String	NEW_EQUIPMENT_CONTAINER;
	/** Used by {@link CSMenuKeys}. */
	public static String	ABOUT;
	/** Used by {@link CSMenuKeys}. */
	public static String	RELEASE_NOTES;
	/** Used by {@link CSMenuKeys}. */
	public static String	TODO_LIST;
	/** Used by {@link CSMenuKeys}. */
	public static String	USERS_MANUAL;
	/** Used by {@link CSMenuKeys}. */
	public static String	LICENSE;
	/** Used by {@link CSMenuKeys}. */
	public static String	WEB_SITE;
	/** Used by {@link CSMenuKeys}. */
	public static String	MAILING_LISTS;

	/** Used by {@link CSNamer}. */
	public static String	NAME_TITLE;
	/** Used by {@link CSNamer}. */
	public static String	CANCEL_REST;
	/** Used by {@link CSNamer}. */
	public static String	CANCEL;
	/** Used by {@link CSNamer}. */
	public static String	APPLY;
	/** Used by {@link CSNamer}. */
	public static String	ONE_REMAINING;
	/** Used by {@link CSNamer}. */
	public static String	REMAINING;

	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_TITLE;
	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_SHEET;
	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_TEMPLATE;
	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_ADVANTAGE_LIST;
	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_SKILL_LIST;
	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_SPELL_LIST;
	/** Used by {@link CSOpenAccessoryPanel}. */
	public static String	NEW_EQUIPMENT_LIST;

	/** Used by {@link CSOutline}. */
	public static String	CLEAR_UNDO;
	/** Used by {@link CSOutline}. */
	public static String	DUPLICATE_UNDO;

	/** Used by {@link CSWindow}. */
	public static String	FILE_MENU;
	/** Used by {@link CSWindow}. */
	public static String	EDIT_MENU;
	/** Used by {@link CSWindow}. */
	public static String	ITEM_MENU;
	/** Used by {@link CSWindow}. */
	public static String	DATA_MENU;
	/** Used by {@link CSWindow}. */
	public static String	WINDOW_MENU;
	/** Used by {@link CSWindow}. */
	public static String	HELP_MENU;
	/** Used by {@link CSWindow}. */
	public static String	SAVE_CHANGES;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
