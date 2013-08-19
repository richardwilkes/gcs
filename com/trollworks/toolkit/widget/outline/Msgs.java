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

package com.trollworks.toolkit.widget.outline;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKOutline}. */
	public static String	COLUMN_MENU_TITLE;
	/** Used by {@link TKOutline}. */
	public static String	SORT_UNDO_TITLE;
	/** Used by {@link TKOutline}. */
	public static String	ROW_DROP_UNDO_TITLE;

	/** Used by {@link TKOutline} and {@link TKOutlineHeaderCM}. */
	public static String	SHOW_ALL_COLUMNS_TITLE;
	/** Used by {@link TKOutline} and {@link TKOutlineHeaderCM}. */
	public static String	RESET_COLUMNS_TITLE;

	/** Used by {@link TKOutlineHeaderCM}. */
	public static String	HIDE_COLUMN_TITLE;
	/** Used by {@link TKOutlineHeaderCM}. */
	public static String	SHOW_COLUMN_TITLE;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
