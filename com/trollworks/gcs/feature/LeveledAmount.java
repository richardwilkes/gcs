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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.text.NumberUtils;

import java.io.IOException;

/** Manages a leveled amount. */
public class LeveledAmount {
	/** The "per level" attribute. */
	public static final String	ATTRIBUTE_PER_LEVEL	= "per_level";	//$NON-NLS-1$
	private boolean				mPerLevel;
	private int					mLevel;
	private double				mAmount;
	private boolean				mInteger;

	/**
	 * Creates a new leveled amount.
	 * 
	 * @param amount The initial amount.
	 */
	public LeveledAmount(double amount) {
		mPerLevel = false;
		mLevel = 0;
		mAmount = amount;
		mInteger = false;
	}

	/**
	 * Creates a new leveled amount.
	 * 
	 * @param amount The initial amount.
	 */
	public LeveledAmount(int amount) {
		this((double) amount);
		mInteger = true;
	}

	/**
	 * Creates a new leveled amount.
	 * 
	 * @param other A {@link LeveledAmount} to clone.
	 */
	public LeveledAmount(LeveledAmount other) {
		mPerLevel = other.mPerLevel;
		mLevel = other.mLevel;
		mAmount = other.mAmount;
		mInteger = other.mInteger;
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof LeveledAmount) {
			LeveledAmount other = (LeveledAmount) obj;

			return mPerLevel == other.mPerLevel && mInteger == other.mInteger && mLevel == other.mLevel && mAmount == other.mAmount;
		}
		return false;
	}

	/**
	 * @param reader The reader to load data from.
	 * @throws IOException
	 */
	public void load(XMLReader reader) throws IOException {
		mPerLevel = reader.isAttributeSet(ATTRIBUTE_PER_LEVEL);
		mAmount = reader.readDouble(0.0);
	}

	/**
	 * Saves this object as XML to a stream.
	 * 
	 * @param out The XML writer to use.
	 * @param tag The tag to use.
	 */
	public void save(XMLWriter out, String tag) {
		if (isPerLevel()) {
			if (mInteger) {
				out.simpleTagWithAttribute(tag, getIntegerAmount(), ATTRIBUTE_PER_LEVEL, mPerLevel);
			} else {
				out.simpleTagWithAttribute(tag, mAmount, ATTRIBUTE_PER_LEVEL, mPerLevel);
			}
		} else {
			if (mInteger) {
				out.simpleTag(tag, getIntegerAmount());
			} else {
				out.simpleTag(tag, mAmount);
			}
		}
	}

	/** @return Whether the amount should be applied per level. */
	public boolean isPerLevel() {
		return mPerLevel;
	}

	/** @param perLevel Whether the amount should be applied per level. */
	public void setPerLevel(boolean perLevel) {
		mPerLevel = perLevel;
	}

	/** @return The current level to use. */
	public int getLevel() {
		return mLevel;
	}

	/** @param level The current level to use. */
	public void setLevel(int level) {
		mLevel = level;
	}

	/** @return Whether this is an integer representation only. */
	public boolean isIntegerOnly() {
		return mInteger;
	}

	/** @param integerOnly Whether this is an integer representation only. */
	public void setIntegerOnly(boolean integerOnly) {
		mInteger = integerOnly;
		if (mInteger) {
			mAmount = Math.round(mAmount);
		}
	}

	/** @return The amount. */
	public double getAmount() {
		return mAmount;
	}

	/** @return The amount. */
	public int getIntegerAmount() {
		return (int) Math.round(mAmount);
	}

	/** @return The amount, as a {@link String}. */
	public String getAmountAsString() {
		if (mInteger) {
			return NumberUtils.format(getIntegerAmount(), true);
		}
		return NumberUtils.format(mAmount, true);
	}

	/** @param amount The amount. */
	public void setAmount(double amount) {
		if (mInteger) {
			mAmount = Math.round(amount);
		} else {
			mAmount = amount;
		}
	}

	/** @param amount The amount. */
	public void setAmount(int amount) {
		mAmount = amount;
	}

	/** @return The amount, adjusted for level, if requested. */
	public double getAdjustedAmount() {
		double amt = mAmount;

		if (isPerLevel()) {
			amt *= getLevel();
		}
		return amt;
	}

	/** @return The amount, adjusted for level, if requested. */
	public int getIntegerAdjustedAmount() {
		int amt = getIntegerAmount();

		if (isPerLevel()) {
			amt *= getLevel();
		}
		return amt;
	}
}
