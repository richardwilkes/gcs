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

import java.io.IOException;

/** Manages numeric comparison criteria. */
public abstract class CMNumericCriteria {
	/** The comparison attribute. */
	public static final String		ATTRIBUTE_COMPARE	= "compare";	//$NON-NLS-1$
	private CMNumericCompareType	mType;

	/**
	 * Creates a new numeric comparison.
	 * 
	 * @param type The {@link CMNumericCompareType} to use.
	 */
	public CMNumericCriteria(CMNumericCompareType type) {
		setType(type);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMNumericCriteria) {
			return mType == ((CMNumericCriteria) obj).mType;
		}
		return false;
	}

	/**
	 * Loads data.
	 * 
	 * @param reader The reader to load data from.
	 * @throws IOException
	 */
	@SuppressWarnings("unused") public void load(TKXMLReader reader) throws IOException {
		setType((CMNumericCompareType) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_COMPARE), CMNumericCompareType.values(), CMNumericCompareType.AT_LEAST));
	}

	/**
	 * Saves this object as XML to a stream.
	 * 
	 * @param out The XML writer to use.
	 * @param tag The tag to use.
	 */
	public void save(TKXMLWriter out, String tag) {
		out.simpleTagWithAttribute(tag, getQualifierAsString(false), ATTRIBUTE_COMPARE, mType.name().toLowerCase());
	}

	/**
	 * @param allowAdornments Whether extras, such as "lbs." can be appended to the text.
	 * @return The numeric qualifier, as a {@link String}.
	 */
	public abstract String getQualifierAsString(boolean allowAdornments);

	/** @return The type of comparison to make. */
	public CMNumericCompareType getType() {
		return mType;
	}

	/** @param type The type of comparison to make. */
	public void setType(CMNumericCompareType type) {
		mType = type;
	}

	@Override public String toString() {
		return toString(Msgs.IS_PREFIX);
	}

	/**
	 * @param prefix A prefix to place before the description.
	 * @return A formatted description of this object.
	 */
	public String toString(String prefix) {
		return mType.format(prefix, getQualifierAsString(true));
	}
}
