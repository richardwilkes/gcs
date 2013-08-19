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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.criteria;

import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.xml.XMLReader;

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
