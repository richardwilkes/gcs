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

import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** Describes a bonus. */
public abstract class CMBonus implements CMFeature {
	/** The "amount" tag. */
	public static final String	TAG_AMOUNT	= "amount"; //$NON-NLS-1$
	private CMLeveledAmount		mAmount;

	/**
	 * Creates a new bonus.
	 * 
	 * @param amount The initial amount.
	 */
	public CMBonus(double amount) {
		mAmount = new CMLeveledAmount(amount);
	}

	/**
	 * Creates a new bonus.
	 * 
	 * @param amount The initial amount.
	 */
	public CMBonus(int amount) {
		mAmount = new CMLeveledAmount(amount);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public CMBonus(CMBonus other) {
		mAmount = new CMLeveledAmount(other.mAmount);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMBonus) {
			return mAmount.equals(((CMBonus) obj).mAmount);
		}
		return false;
	}

	/**
	 * @param reader The XML reader to use.
	 * @throws IOException
	 */
	protected final void load(TKXMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				loadSelf(reader);
			}
		} while (reader.withinMarker(marker));
	}

	/**
	 * @param reader The XML reader to use.
	 * @throws IOException
	 */
	protected void loadSelf(TKXMLReader reader) throws IOException {
		String tag = reader.getName();

		if (TAG_AMOUNT.equals(tag)) {
			mAmount.load(reader);
		} else {
			reader.skipTag(tag);
		}
	}

	/**
	 * Saves the bonus base information.
	 * 
	 * @param out The XML writer to use..
	 */
	public void saveBase(TKXMLWriter out) {
		mAmount.save(out, TAG_AMOUNT);
	}

	/** @return The leveled amount. */
	public CMLeveledAmount getAmount() {
		return mAmount;
	}

	public void fillWithNameableKeys(HashSet<String> set) {
		// Nothing to do.
	}

	public void applyNameableKeys(HashMap<String, String> map) {
		// Nothing to do.
	}
}
