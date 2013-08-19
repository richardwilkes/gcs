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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKFileDialog}. */
	public static String	OPEN_DIALOG_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	SAVE_DIALOG_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	OPEN_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	SAVE_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	FILTER_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	FILENAME_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	WRITE_PERMISSION_ERROR;
	/** Used by {@link TKFileDialog}. */
	public static String	IS_FOLDER_ERROR;
	/** Used by {@link TKFileDialog}. */
	public static String	OVERWRITE_CONFIRMATION;
	/** Used by {@link TKFileDialog}. */
	public static String	OVERWRITE_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	UP_FOLDER_TOOLTIP;
	/** Used by {@link TKFileDialog}. */
	public static String	NEW_FOLDER_TOOLTIP;
	/** Used by {@link TKFileDialog}. */
	public static String	HOME_TOOLTIP;
	/** Used by {@link TKFileDialog}. */
	public static String	FAVORITES;
	/** Used by {@link TKFileDialog}. */
	public static String	RECENT_FOLDERS;
	/** Used by {@link TKFileDialog}. */
	public static String	ADD_TO_FAVORITES;
	/** Used by {@link TKFileDialog}. */
	public static String	REMOVE_FROM_FAVORITES;
	/** Used by {@link TKFileDialog}. */
	public static String	REMOVAL_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	REMOVAL_DIALOG_TITLE;
	/** Used by {@link TKFileDialog}. */
	public static String	DEFAULT_NEW_FOLDER_NAME;
	/** Used by {@link TKFileDialog}. */
	public static String	NEW_FOLDER_PROMPT;
	/** Used by {@link TKFileDialog}. */
	public static String	UNABLE_TO_CREATE_NEW_FOLDER;

	/** Used by {@link TKFileDialog} and {@link TKOptionDialog}. */
	public static String	CANCEL_TITLE;

	/** Used by {@link TKOptionDialog}. */
	public static String	OK_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	YES_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	NO_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	WARNING_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	ERROR_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	MESSAGE_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	CONFIRMATION_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	RESPONSE_TITLE;
	/** Used by {@link TKOptionDialog}. */
	public static String	DONT_SHOW_AGAIN;

	/** Used by {@link TKOpenManager}. */
	public static String	OPEN_FAILURE;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
