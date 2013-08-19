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

package com.trollworks.toolkit.qa;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKQAMenu}. */
	public static String	MENU_TITLE;
	/** Used by {@link TKQAMenu}. */
	public static String	MONITOR_UNDO;
	/** Used by {@link TKQAMenu}. */
	public static String	DIAGNOSE_LOAD_SAVE;
	/** Used by {@link TKQAMenu}. */
	public static String	DOUBLE_BUFFERED;
	/** Used by {@link TKQAMenu}. */
	public static String	SHOW_DRAWING;
	/** Used by {@link TKQAMenu}. */
	public static String	SHOW_PANELS_WITHOUT_TOOLTIPS;
	/** Used by {@link TKQAMenu}. */
	public static String	ANTIALIASED_FONTS;
	/** Used by {@link TKQAMenu}. */
	public static String	FRACTIONAL_FONT_METRICS;
	/** Used by {@link TKQAMenu}. */
	public static String	HIGH_QUALITY_RENDERING;
	/** Used by {@link TKQAMenu}. */
	public static String	NORMALIZE_STROKES;
	/** Used by {@link TKQAMenu}. */
	public static String	ANTIALIASING;
	/** Used by {@link TKQAMenu}. */
	public static String	HIGH_QUALITY_COLOR_RENDERING;
	/** Used by {@link TKQAMenu}. */
	public static String	DITHER_WHEN_NEEDED;
	/** Used by {@link TKQAMenu}. */
	public static String	INTERPOLATION;
	/** Used by {@link TKQAMenu}. */
	public static String	BICUBIC;
	/** Used by {@link TKQAMenu}. */
	public static String	BILINEAR;
	/** Used by {@link TKQAMenu}. */
	public static String	NEAREST_NEIGHBOR;
	/** Used by {@link TKQAMenu}. */
	public static String	HIGH_QUALITY_ALPHA_INTERPOLATION;
	/** Used by {@link TKQAMenu}. */
	public static String	SET_MAXIMUM_QUALITY;
	/** Used by {@link TKQAMenu}. */
	public static String	SET_MINIMUM_QUALITY;
	/** Used by {@link TKQAMenu}. */
	public static String	SHOW_RENDERING_HINTS;
	/** Used by {@link TKQAMenu}. */
	public static String	CURRENT_WINDOW_SIZE_FORMAT;

	/** Used by {@link TKUndoMonitorWindow}. */
	public static String	TITLE_FORMAT;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
