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

import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;

/** Manages string comparison criteria. */
public class StringCriteria {
	private static final String	ATTRIBUTE_COMPARE	= "compare";	//$NON-NLS-1$
	private StringCompareType	mType;
	private String				mQualifier;

	/**
	 * Creates a new string comparison.
	 * 
	 * @param type The type of comparison.
	 * @param qualifier The qualifier to match against.
	 */
	public StringCriteria(StringCompareType type, String qualifier) {
		setType(type);
		setQualifier(qualifier);
	}

	/**
	 * Creates a new string comparison.
	 * 
	 * @param other A {@link StringCriteria} to clone.
	 */
	public StringCriteria(StringCriteria other) {
		mType = other.mType;
		mQualifier = other.mQualifier;
	}

	@Override
	public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof StringCriteria && super.equals(obj)) {
			StringCriteria other = (StringCriteria) obj;

			return mType == other.mType && mQualifier.equalsIgnoreCase(other.mQualifier);
		}
		return false;
	}

	/**
	 * @param reader The reader to load data from.
	 * @throws IOException
	 */
	public void load(XMLReader reader) throws IOException {
		setType(StringCompareType.get(reader.getAttribute(ATTRIBUTE_COMPARE)));
		setQualifier(reader.readText());
	}

	/**
	 * Saves this object as XML to a stream.
	 * 
	 * @param out The XML writer to use.
	 * @param tag The tag to use.
	 */
	public void save(XMLWriter out, String tag) {
		out.simpleTagWithAttribute(tag, mQualifier, ATTRIBUTE_COMPARE, mType.toString());
	}

	/** @return The type of comparison to make. */
	public StringCompareType getType() {
		return mType;
	}

	/** @param type The type of comparison to make. */
	public void setType(StringCompareType type) {
		mType = type;
	}

	/** @return The qualifier to match against. */
	public String getQualifier() {
		return mQualifier;
	}

	/** @param qualifier The qualifier to match against. */
	public void setQualifier(String qualifier) {
		mQualifier = qualifier != null ? qualifier : ""; //$NON-NLS-1$
	}

	/**
	 * @param data The data to match against.
	 * @return Whether the data matches this criteria.
	 */
	public boolean matches(String data) {
		return mType.matches(mQualifier, data);
	}

	@Override
	public String toString() {
		return mType.describe(mQualifier);
	}
}
