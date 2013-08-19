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

package com.trollworks.toolkit.print;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link TKInkChromaticity}. */
	public static String	COLOR;
	/** Used by {@link TKInkChromaticity}. */
	public static String	MONOCHROME;

	/** Used by {@link TKPageOrientation}. */
	public static String	LANDSCAPE;
	/** Used by {@link TKPageOrientation}. */
	public static String	PORTRAIT;
	/** Used by {@link TKPageOrientation}. */
	public static String	REVERSE_LANDSCAPE;
	/** Used by {@link TKPageOrientation}. */
	public static String	REVERSE_PORTRAIT;

	/** Used by {@link TKQuality}. */
	public static String	HIGH;
	/** Used by {@link TKQuality}. */
	public static String	NORMAL;
	/** Used by {@link TKQuality}. */
	public static String	DRAFT;

	/** Used by {@link TKPageSides}. */
	public static String	SINGLE;
	/** Used by {@link TKPageSides}. */
	public static String	DUPLEX;
	/** Used by {@link TKPageSides}. */
	public static String	TUMBLE;

	/** Used by {@link TKPrintManager}. */
	public static String	PRINTING_FAILED;
	/** Used by {@link TKPrintManager}. */
	public static String	UNABLE_TO_SWITCH_PRINTERS;
	/** Used by {@link TKPrintManager}. */
	public static String	NO_PRINTER_AVAILABLE;
	/** Used by {@link TKPrintManager}. */
	public static String	PAGE_SETUP_TITLE;
	/** Used by {@link TKPrintManager}. */
	public static String	PRINT_TITLE;

	/** Used by {@link TKPageSetupPanel}. */
	public static String	PRINTER;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	ORIENTATION;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	PAPER_TYPE;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	MARGINS;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	CHROMATICITY;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	SIDES;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	NUMBER_UP;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	QUALITY;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	RESOLUTION;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	DPI;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	PAGE_SETTINGS;
	/** Used by {@link TKPageSetupPanel}. */
	public static String	QUALITY_SETTINGS;

	/** Used by {@link TKPrintPanel}. */
	public static String	COPIES;
	/** Used by {@link TKPrintPanel}. */
	public static String	PAGE_RANGE;
	/** Used by {@link TKPrintPanel}. */
	public static String	ALL;
	/** Used by {@link TKPrintPanel}. */
	public static String	PAGES;
	/** Used by {@link TKPrintPanel}. */
	public static String	TO;
	/** Used by {@link TKPrintPanel}. */
	public static String	PRINT_JOB_SETTINGS;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
