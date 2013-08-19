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

/** The attribute affected by a {@link CMAttributeBonus}. */
public enum CMBonusAttributeType {
	/** The ST attribute. */
	ST(Msgs.ST),
	/** The DX attribute. */
	DX(Msgs.DX),
	/** The IQ attribute. */
	IQ(Msgs.IQ),
	/** The HT attribute. */
	HT(Msgs.HT),
	/** The Will attribute. */
	WILL(Msgs.WILL),
	/** The Perception attribute. */
	PERCEPTION(Msgs.PERCEPTION),
	/** The Vision attribute. */
	VISION(Msgs.VISION),
	/** The Hearing attribute. */
	HEARING(Msgs.HEARING),
	/** The TasteSmell attribute. */
	TASTE_SMELL(Msgs.TASTE_SMELL),
	/** The Touch attribute. */
	TOUCH(Msgs.TOUCH),
	/** The Dodge attribute. */
	DODGE(Msgs.DODGE),
	/** The Dodge attribute. */
	PARRY(Msgs.PARRY),
	/** The Dodge attribute. */
	BLOCK(Msgs.BLOCK),
	/** The Speed attribute. */
	SPEED(Msgs.SPEED) {
		@Override public boolean isIntegerOnly() {
			return false;
		}
	},
	/** The Move attribute. */
	MOVE(Msgs.MOVE),
	/** The FP attribute. */
	FP(Msgs.FP),
	/** The HP attribute. */
	HP(Msgs.HP),
	/** The size modifier attribute. */
	SM(Msgs.SM);

	private String	mTitle;
	private String	mTag;

	private CMBonusAttributeType(String title) {
		mTitle = title;
		mTag = name();
		if (mTag.length() > 2) {
			mTag = mTag.toLowerCase();
		}
	}

	/** @return The presenation name. */
	public String getPresentationName() {
		String name = name();
		if (name.length() > 2) {
			name = Character.toUpperCase(name.charAt(0)) + name.substring(1).toLowerCase();
		}
		return name;
	}

	@Override public String toString() {
		return mTitle;
	}

	/** @return <code>true</code> if only integer values are permitted. */
	public boolean isIntegerOnly() {
		return true;
	}

	/** @return The XML tag to use for this {@link CMBonusAttributeType}. */
	public String getXMLTag() {
		return mTag;
	}
}
