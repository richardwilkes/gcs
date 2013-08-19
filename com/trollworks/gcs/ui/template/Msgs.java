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

package com.trollworks.gcs.ui.template;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSTemplate}. */
	public static String	ADVANTAGES;
	/** Used by {@link CSTemplate}. */
	public static String	SKILLS;
	/** Used by {@link CSTemplate}. */
	public static String	SPELLS;
	/** Used by {@link CSTemplate}. */
	public static String	EQUIPMENT;
	/** Used by {@link CSTemplate}. */
	public static String	NOTES;

	/** Used by {@link CSTemplateOpener}. */
	public static String	TEMPLATES;

	/** Used by {@link CSTemplateWindow}. */
	public static String	UNTITLED;
	/** Used by {@link CSTemplateWindow}. */
	public static String	ADD_ROWS;
	/** Used by {@link CSTemplateWindow}. */
	public static String	UNDO;
	/** Used by {@link CSTemplateWindow}. */
	public static String	SAVE_ERROR;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
