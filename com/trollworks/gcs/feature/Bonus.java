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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** Describes a bonus. */
public abstract class Bonus implements Feature {
	/** The "amount" tag. */
	public static final String	TAG_AMOUNT	= "amount"; //$NON-NLS-1$
	private LeveledAmount		mAmount;

	/**
	 * Creates a new bonus.
	 * 
	 * @param amount The initial amount.
	 */
	public Bonus(double amount) {
		mAmount = new LeveledAmount(amount);
	}

	/**
	 * Creates a new bonus.
	 * 
	 * @param amount The initial amount.
	 */
	public Bonus(int amount) {
		mAmount = new LeveledAmount(amount);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public Bonus(Bonus other) {
		mAmount = new LeveledAmount(other.mAmount);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof Bonus) {
			return mAmount.equals(((Bonus) obj).mAmount);
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	/** @param reader The XML reader to use. */
	protected final void load(XMLReader reader) throws IOException {
		String marker = reader.getMarker();

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				loadSelf(reader);
			}
		} while (reader.withinMarker(marker));
	}

	/** @param reader The XML reader to use. */
	protected void loadSelf(XMLReader reader) throws IOException {
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
	public void saveBase(XMLWriter out) {
		mAmount.save(out, TAG_AMOUNT);
	}

	/** @return The leveled amount. */
	public LeveledAmount getAmount() {
		return mAmount;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		// Nothing to do.
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		// Nothing to do.
	}
}
