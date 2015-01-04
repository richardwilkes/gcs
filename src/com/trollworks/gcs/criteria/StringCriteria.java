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
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.text.Enums;

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
		if (obj == this) {
			return true;
		}
		if (obj instanceof StringCriteria) {
			StringCriteria sc = (StringCriteria) obj;
			return mType == sc.mType && mQualifier.equalsIgnoreCase(sc.mQualifier);
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	/**
	 * @param reader The reader to load data from.
	 */
	public void load(XMLReader reader) throws IOException {
		setType(Enums.extract(reader.getAttribute(ATTRIBUTE_COMPARE), StringCompareType.values()));
		setQualifier(reader.readText());
	}

	/**
	 * Saves this object as XML to a stream.
	 *
	 * @param out The XML writer to use.
	 * @param tag The tag to use.
	 */
	public void save(XMLWriter out, String tag) {
		out.simpleTagWithAttribute(tag, mQualifier, ATTRIBUTE_COMPARE, Enums.toId(mType));
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
