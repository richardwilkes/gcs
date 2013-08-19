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

package com.trollworks.gcs.ui.advantage;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSAdvantageColumnID}. */
	public static String	DESCRIPTION;
	/** Used by {@link CSAdvantageColumnID}. */
	public static String	DESCRIPTION_TOOLTIP;
	/** Used by {@link CSAdvantageColumnID}. */
	public static String	POINTS;
	/** Used by {@link CSAdvantageColumnID}. */
	public static String	POINTS_TOOLTIP;
	/** Used by {@link CSAdvantageColumnID}. */
	public static String	TYPE_TOOLTIP;
	/** Used by {@link CSAdvantageColumnID}. */
	public static String	REFERENCE;

	/** Used by {@link CSAdvantageEditor}. */
	public static String	NAME;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	NAME_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	TOTAL_POINTS;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	TOTAL_POINTS_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	BASE_POINTS;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	BASE_POINTS_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	LEVEL_POINTS;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	LEVEL_POINTS_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	LEVEL;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	LEVEL_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	NOTES;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	NOTES_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	EDTIOR_TYPE_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	CONTAINER_TYPE;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	CONTAINER_TYPE_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	EDITOR_REFERENCE_TOOLTIP;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	NO_LEVELS;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	HAS_LEVELS;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	MENTAL;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	PHYSICAL;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	SOCIAL;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	EXOTIC;
	/** Used by {@link CSAdvantageEditor}. */
	public static String	SUPERNATURAL;

	/** Used by {@link CSAdvantageColumnID} and {@link CSAdvantageEditor}. */
	public static String	TYPE;
	/** Used by {@link CSAdvantageColumnID} and {@link CSAdvantageEditor}. */
	public static String	REFERENCE_TOOLTIP;

	/** Used by {@link CSAdvantageListWindow}. */
	public static String	UNTITLED;

	/** Used by {@link CSAdvantageOutline}. */
	public static String	INCREMENT;
	/** Used by {@link CSAdvantageOutline}. */
	public static String	DECREMENT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
