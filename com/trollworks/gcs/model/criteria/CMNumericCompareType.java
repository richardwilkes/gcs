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

package com.trollworks.gcs.model.criteria;

import java.text.MessageFormat;

/** The allowed numeric comparison types. */
public enum CMNumericCompareType {
	/** The comparison for "is". */
	IS(Msgs.IS, Msgs.IS_DESCRIPTION2, Msgs.IS_DESCRIPTION),
	/** The comparison for "is at least". */
	AT_LEAST(Msgs.AT_LEAST, Msgs.AT_LEAST_DESCRIPTION2, Msgs.AT_LEAST_DESCRIPTION),
	/** The comparison for "is at most". */
	AT_MOST(Msgs.AT_MOST, Msgs.AT_MOST_DESCRIPTION2, Msgs.AT_MOST_DESCRIPTION);

	private String	mTitle;
	private String	mDescription;
	private String	mDescriptionFormat;

	private CMNumericCompareType(String title, String description, String descriptionFormat) {
		mTitle = title;
		mDescription = description;
		mDescriptionFormat = descriptionFormat;
	}

	@Override public String toString() {
		return mTitle;
	}

	/** @return A description of this object. */
	public String getDescription() {
		return mDescription;
	}

	/**
	 * @param prefix A prefix to place before the description.
	 * @param qualifier The qualifier to use.
	 * @return A formatted description of this object.
	 */
	public String format(String prefix, String qualifier) {
		return MessageFormat.format(mDescriptionFormat, prefix, qualifier);
	}
}
