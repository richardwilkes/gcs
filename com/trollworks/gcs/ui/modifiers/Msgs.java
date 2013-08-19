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

package com.trollworks.gcs.ui.modifiers;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CSModifierEditor}. */
	public static String	NAME;
	/** Used by {@link CSModifierEditor}. */
	public static String	NAME_TOOLTIP;
	/** Used by {@link CSModifierEditor}. */
	public static String	NOTES;
	/** Used by {@link CSModifierEditor}. */
	public static String	NOTES_TOOLTIP;
	/** Used by {@link CSModifierEditor}. */
	public static String	NAME_CANNOT_BE_EMPTY;
	/** Used by {@link CSModifierEditor}. */
	public static String	COST;
	/** Used by {@link CSModifierEditor}. */
	public static String	COST_TOOLTIP;
	/** Used by {@link CSModifierEditor} */
	public static String	LEVELS;
	/** Used by {@link CSModifierEditor}. */
	public static String	LEVELS_TOOLTIP;
	/** Used by {@link CSModifierEditor}. */
	public static String	TOTAL_COST;
	/** Used by {@link CSModifierEditor}. */
	public static String	TOTAL_COST_TOOLTIP;
	/** Used by {@link CSModifierEditor}. */
	public static String	NO_LEVELS;
	/** Used by {@link CSModifierEditor}. */
	public static String	HAS_LEVELS;

	/** Used by {@link CSModifierListEditor}. */
	public static String	MODIFIERS;

	/** Used by {@link CSModifierColumnID}. */
	public static String	DESCRIPTION;
	/** Used by {@link CSModifierColumnID}. */
	public static String	DESCRIPTION_TOOLTIP;
	/** Used by {@link CSModifierColumnID}. */
	public static String	COST_MODIFIER;
	/** Used by {@link CSModifierColumnID}. */
	public static String	COST_MODIFIER_TOOLTIP;
	
	/** Used by {@link CSModifierEditor} and {@link CSModifierColumnID}. */
	public static String	REFERENCE;
	/** Used by {@link CSModifierEditor} and {@link CSModifierColumnID}. */
	public static String	REFERENCE_TOOLTIP;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
