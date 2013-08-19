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

package com.trollworks.gcs.model.modifier;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CMModifier}. */
	public static String	DEFAULT_NAME;
	/** Used by {@link CMModifier}. */
	public static String	MODIFIER_TYPE;
	/** Used by {@link CMModifier}. */
	public static String	READ_ONLY;

	/** Used by {@link CMAffects}. */
	public static String	TOTAL;
	/** Used by {@link CMAffects}. */
	public static String	BASE_ONLY;
	/** Used by {@link CMAffects}. */
	public static String	LEVELS_ONLY;
	/** Used by {@link CMAffects}. */
	public static String	TOTAL_SHORT;
	/** Used by {@link CMAffects}. */
	public static String	BASE_ONLY_SHORT;
	/** Used by {@link CMAffects}. */
	public static String	LEVELS_ONLY_SHORT;

	/** Used by {@link CMCostType}. */
	public static String	PERCENTAGE;
	/** Used by {@link CMCostType}. */
	public static String	POINTS;
	/** Used by {@link CMCostType}. */
	public static String	MULTIPLIER;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
