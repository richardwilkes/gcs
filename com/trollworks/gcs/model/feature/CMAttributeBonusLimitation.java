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

package com.trollworks.gcs.model.feature;

/** The limitations applicable to a {@link CMAttributeBonus}. */
public enum CMAttributeBonusLimitation {
	/** No limitation. */
	NONE(Msgs.NONE),
	/** Striking only. */
	STRIKING_ONLY(Msgs.STRIKING_ONLY),
	/** Lifting only */
	LIFTING_ONLY(Msgs.LIFTING_ONLY);

	private String mTitle;

	private CMAttributeBonusLimitation(String title) {
		mTitle = title;
	}

	@Override public String toString() {
		return mTitle;
	}
}
