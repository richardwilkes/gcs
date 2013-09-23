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

package com.trollworks.gcs.feature;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;

@Localized({
				@LS(key = "ST", msg = "to ST"),
				@LS(key = "DX", msg = "to DX"),
				@LS(key = "IQ", msg = "to IQ"),
				@LS(key = "HT", msg = "to HT"),
				@LS(key = "WILL", msg = "to will"),
				@LS(key = "FRIGHT_CHECK", msg = "to fright checks"),
				@LS(key = "PERCEPTION", msg = "to perception"),
				@LS(key = "VISION", msg = "to vision"),
				@LS(key = "HEARING", msg = "to hearing"),
				@LS(key = "TASTE_SMELL", msg = "to taste & smell"),
				@LS(key = "TOUCH", msg = "to touch"),
				@LS(key = "DODGE", msg = "to dodge"),
				@LS(key = "PARRY", msg = "to parry"),
				@LS(key = "BLOCK", msg = "to block"),
				@LS(key = "SPEED", msg = "to basic speed"),
				@LS(key = "MOVE", msg = "to basic move"),
				@LS(key = "FP", msg = "to FP"),
				@LS(key = "HP", msg = "to HP"),
				@LS(key = "SM", msg = "to size modifier"),
})
/** The attribute affected by a {@link AttributeBonus}. */
public enum BonusAttributeType {
	/** The ST attribute. */
	ST,
	/** The DX attribute. */
	DX,
	/** The IQ attribute. */
	IQ,
	/** The HT attribute. */
	HT,
	/** The Will attribute. */
	WILL,
	/** The Fright Check attribute. */
	FRIGHT_CHECK,
	/** The Perception attribute. */
	PERCEPTION,
	/** The Vision attribute. */
	VISION,
	/** The Hearing attribute. */
	HEARING,
	/** The TasteSmell attribute. */
	TASTE_SMELL,
	/** The Touch attribute. */
	TOUCH,
	/** The Dodge attribute. */
	DODGE,
	/** The Dodge attribute. */
	PARRY,
	/** The Dodge attribute. */
	BLOCK,
	/** The Speed attribute. */
	SPEED {
		@Override
		public boolean isIntegerOnly() {
			return false;
		}
	},
	/** The Move attribute. */
	MOVE,
	/** The FP attribute. */
	FP,
	/** The HP attribute. */
	HP,
	/** The size modifier attribute. */
	SM;

	@Override
	public String toString() {
		return BonusAttributeType_LS.toString(this);
	}

	private String	mTag;

	private BonusAttributeType() {
		mTag = name();
		if (mTag.length() > 2) {
			mTag = mTag.toLowerCase();
		}
	}

	/** @return The presentation name. */
	public String getPresentationName() {
		String name = name();
		if (name.length() > 2) {
			name = Character.toUpperCase(name.charAt(0)) + name.substring(1).toLowerCase();
		}
		return name;
	}

	/** @return <code>true</code> if only integer values are permitted. */
	@SuppressWarnings("static-method")
	public boolean isIntegerOnly() {
		return true;
	}

	/** @return The XML tag to use for this {@link BonusAttributeType}. */
	public String getXMLTag() {
		return mTag;
	}
}
