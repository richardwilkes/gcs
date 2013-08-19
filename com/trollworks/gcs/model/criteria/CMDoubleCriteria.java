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

import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.utility.units.TKWeightUnits;

import java.io.IOException;

/** Manages double comparison criteria. */
public class CMDoubleCriteria extends CMNumericCriteria {
	private static final String	ATTRIBUTE_UNITS	= "units";	//$NON-NLS-1$
	private double				mQualifier;
	private boolean				mIsWeight;

	/**
	 * Creates a new double comparison.
	 * 
	 * @param type One of {@link #IS}, {@link #AT_LEAST}, or {@link #NO_MORE_THAN}.
	 * @param qualifier The qualifier to match against.
	 * @param isWeight Whether this number represents a weight.
	 */
	public CMDoubleCriteria(String type, double qualifier, boolean isWeight) {
		super(type);
		setQualifier(qualifier);
		mIsWeight = isWeight;
	}

	/**
	 * Creates a new double comparison.
	 * 
	 * @param other A {@link CMDoubleCriteria} to clone.
	 */
	public CMDoubleCriteria(CMDoubleCriteria other) {
		super(other.getType());
		mQualifier = other.mQualifier;
		mIsWeight = other.mIsWeight;
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMDoubleCriteria && super.equals(obj)) {
			CMDoubleCriteria other = (CMDoubleCriteria) obj;

			return mIsWeight == other.mIsWeight && mQualifier == other.mQualifier;
		}
		return false;
	}

	@Override public void load(TKXMLReader reader) throws IOException {
		super.load(reader);
		if (mIsWeight) {
			setQualifier(TKWeightUnits.POUNDS.convert((TKWeightUnits) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), TKWeightUnits.values(), TKWeightUnits.POUNDS), reader.readDouble(0)));
		} else {
			setQualifier(reader.readDouble(0.0));
		}
	}

	@Override public void save(TKXMLWriter out, String tag) {
		if (mIsWeight) {
			out.startTag(tag);
			out.writeAttribute(ATTRIBUTE_COMPARE, getType());
			out.writeAttribute(ATTRIBUTE_UNITS, TKWeightUnits.POUNDS.toString());
			out.finishTag();
			out.writeEncodedData(getQualifierAsString(false));
			out.endTagEOL(tag, false);
		} else {
			super.save(out, tag);
		}
	}

	/** @return Whether this number represents a weight. */
	public boolean isWeight() {
		return mIsWeight;
	}

	/** @return The qualifier to match against. */
	public double getQualifier() {
		return mQualifier;
	}

	@Override public String getQualifierAsString(boolean allowAdornments) {
		if (mIsWeight && allowAdornments) {
			return TKWeightUnits.POUNDS.format(mQualifier);
		}
		return Double.toString(mQualifier);
	}

	/** @param qualifier The qualifier to match against. */
	public void setQualifier(double qualifier) {
		mQualifier = qualifier;
	}

	/**
	 * @param data The data to match against.
	 * @return Whether the data matches this criteria.
	 */
	public boolean matches(double data) {
		String type = getType();

		if (IS == type) {
			return data == mQualifier;
		}
		if (AT_LEAST == type) {
			return data >= mQualifier;
		}
		if (NO_MORE_THAN == type) {
			return data <= mQualifier;
		}
		return false;
	}
}
