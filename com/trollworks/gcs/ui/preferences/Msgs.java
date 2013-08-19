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

package com.trollworks.gcs.ui.preferences;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSFontPreferences}. */
	public static String	FONTS;
	/** Used by {@link CSFontPreferences}. */
	public static String	ANTIALIAS_FONTS;
	/** Used by {@link CSFontPreferences}. */
	public static String	CONTROLS_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	TEXT_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	MENUS_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	MENU_KEYS_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	LABELS_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	FIELDS_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	FIELD_NOTES_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	TECHNIQUE_FIELDS_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	PRIMARY_FOOTER_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	SECONDARY_FOOTER_FONT;
	/** Used by {@link CSFontPreferences}. */
	public static String	NOTES_FONT;

	/** Used by {@link CSGeneralPreferences}. */
	public static String	GENERAL;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DISPLAY_TOOLTIPS;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DISPLAY_TOOLTIPS_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DELAY;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DELAY_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DURATION;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DURATION_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DISPLAY_SPLASH;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	DISPLAY_SPLASH_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	USE_NATIVE_PRINT_DIALOGS;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	USE_NATIVE_PRINT_DIALOGS_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	UNDO_LEVELS_PRE;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	UNDO_LEVELS_POST;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	UNDO_LEVELS_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	BROWSER;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	BROWSER_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	CUSTOM;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	COMMAND;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	COMMAND_TOOLTIP;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	WARNING;
	/** Used by {@link CSGeneralPreferences}. */
	public static String	WARNING_FORMAT;

	/** Used by {@link CSKeystrokeDialog} and {@link CSMenuKeyPreferences}. */
	public static String	NOT_ASSIGNED;

	/** Used by {@link CSKeystrokeDialog}. */
	public static String	PROMPT;
	/** Used by {@link CSKeystrokeDialog}. */
	public static String	ALREADY_ASSIGNED;
	/** Used by {@link CSKeystrokeDialog}. */
	public static String	NO_MODIFIER;

	/** Used by {@link CSMenuKeyPreferences}. */
	public static String	MENU_KEYS;

	/** Used by {@link CSPortraitPreferencePanel}. */
	public static String	PORTRAIT;
	/** Used by {@link CSPortraitPreferencePanel}. */
	public static String	PORTRAIT_TOOLTIP;

	/** Used by {@link CSPreferencesWindow}. */
	public static String	PREFERENCES;
	/** Used by {@link CSPreferencesWindow}. */
	public static String	RESET;

	/** Used by {@link CSSheetPreferences}. */
	public static String	SHEET;
	/** Used by {@link CSSheetPreferences}. */
	public static String	SHEET_DEFAULTS;
	/** Used by {@link CSSheetPreferences}. */
	public static String	PLAYER;
	/** Used by {@link CSSheetPreferences}. */
	public static String	PLAYER_TOOLTIP;
	/** Used by {@link CSSheetPreferences}. */
	public static String	CAMPAIGN;
	/** Used by {@link CSSheetPreferences}. */
	public static String	CAMPAIGN_TOOLTIP;
	/** Used by {@link CSSheetPreferences}. */
	public static String	IMAGE_FILES;
	/** Used by {@link CSSheetPreferences}. */
	public static String	TECH_LEVEL;
	/** Used by {@link CSSheetPreferences}. */
	public static String	TECH_LEVEL_TOOLTIP;
	/** Used by {@link CSSheetPreferences}. */
	public static String	PNG_RESOLUTION;
	/** Used by {@link CSSheetPreferences}. */
	public static String	PNG_RESOLUTION_TOOLTIP;
	/** Used by {@link CSSheetPreferences}. */
	public static String	DPI;

	/** Used by {@link CSGeneralPreferences} and {@link CSSheetPreferences}. */
	public static String	MISCELLANEOUS;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
