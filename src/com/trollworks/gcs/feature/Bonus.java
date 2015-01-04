/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;

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
