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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.criteria;

import static com.trollworks.gcs.criteria.NumericCompareType_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;

import java.text.MessageFormat;

@Localized({
				@LS(key = "IS", msg = "is"),
				@LS(key = "AT_LEAST", msg = "at least"),
				@LS(key = "AT_MOST", msg = "at most"),
				@LS(key = "IS_FORMAT", msg = "{0}\"{1}\""),
				@LS(key = "AT_LEAST_FORMAT", msg = "{0}at least \"{1}\""),
				@LS(key = "AT_MOST_FORMAT", msg = "{0}at most \"{1}\""),
				@LS(key = "IS_DESCRIPTION", msg = "is"),
				@LS(key = "AT_LEAST_DESCRIPTION", msg = "is at least"),
				@LS(key = "AT_MOST_DESCRIPTION", msg = "is at most"),
})
/** The allowed numeric comparison types. */
public enum NumericCompareType {
	/** The comparison for "is". */
	IS(IS_DESCRIPTION, IS_FORMAT),
	/** The comparison for "is at least". */
	AT_LEAST(AT_LEAST_DESCRIPTION, AT_LEAST_FORMAT),
	/** The comparison for "is at most". */
	AT_MOST(AT_MOST_DESCRIPTION, AT_MOST_FORMAT);

	private String	mDescription;
	private String	mDescriptionFormat;

	private NumericCompareType(String description, String descriptionFormat) {
		mDescription = description;
		mDescriptionFormat = descriptionFormat;
	}

	/** @return A description of this object. */
	public String getDescription() {
		return mDescription;
	}

	@Override
	public String toString() {
		return NumericCompareType_LS.toString(this);
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
