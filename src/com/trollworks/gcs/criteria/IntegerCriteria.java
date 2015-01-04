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

package com.trollworks.gcs.criteria;

import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.utility.text.Numbers;

import java.io.IOException;

/** Manages integer comparison criteria. */
public class IntegerCriteria extends NumericCriteria {
	private int	mQualifier;

	/**
	 * Creates a new integer comparison.
	 * 
	 * @param type The {@link NumericCompareType} to use.
	 * @param qualifier The qualifier to match against.
	 */
	public IntegerCriteria(NumericCompareType type, int qualifier) {
		super(type);
		setQualifier(qualifier);
	}

	/**
	 * Creates a new integer comparison.
	 * 
	 * @param other A {@link IntegerCriteria} to clone.
	 */
	public IntegerCriteria(IntegerCriteria other) {
		super(other.getType());
		mQualifier = other.mQualifier;
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof IntegerCriteria && super.equals(obj)) {
			IntegerCriteria criteria = (IntegerCriteria) obj;
			return mQualifier == criteria.mQualifier;
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	@Override
	public void load(XMLReader reader) throws IOException {
		super.load(reader);
		setQualifier(reader.readInteger(0));
	}

	/** @return The qualifier to match against. */
	public int getQualifier() {
		return mQualifier;
	}

	@Override
	public String getQualifierAsString(boolean allowAdornments) {
		return allowAdornments ? Numbers.format(mQualifier) : Integer.toString(mQualifier);
	}

	/** @param qualifier The qualifier to match against. */
	public void setQualifier(int qualifier) {
		mQualifier = qualifier;
	}

	/**
	 * @param data The data to match against.
	 * @return Whether the data matches this criteria.
	 */
	public boolean matches(int data) {
		switch (getType()) {
			case IS:
				return data == mQualifier;
			case AT_LEAST:
			default:
				return data >= mQualifier;
			case AT_MOST:
				return data <= mQualifier;
		}
	}
}
