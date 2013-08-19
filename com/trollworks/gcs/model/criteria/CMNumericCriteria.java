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

import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;

/** Manages numeric comparison criteria. */
public abstract class CMNumericCriteria {
	/** The comparison for "is". */
	public static final String	IS					= "is";		//$NON-NLS-1$
	/** The comparison for "is at least". */
	public static final String	AT_LEAST			= "atLeast";	//$NON-NLS-1$
	/** The comparison for "is no more than". */
	public static final String	NO_MORE_THAN		= "atMost";	//$NON-NLS-1$
	/** The comparison attribute. */
	public static final String	ATTRIBUTE_COMPARE	= "compare";	//$NON-NLS-1$
	private String				mType;

	/**
	 * Creates a new numeric comparison.
	 * 
	 * @param type One of {@link #IS}, {@link #AT_LEAST}, or {@link #NO_MORE_THAN}.
	 */
	public CMNumericCriteria(String type) {
		setType(type);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMNumericCriteria) {
			return mType.equals(((CMNumericCriteria) obj).mType);
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
		setType(reader.getAttribute(ATTRIBUTE_COMPARE));
	}

	/**
	 * Saves this object as XML to a stream.
	 * 
	 * @param out The XML writer to use.
	 * @param tag The tag to use.
	 */
	public void save(TKXMLWriter out, String tag) {
		out.simpleTagWithAttribute(tag, getQualifierAsString(false), ATTRIBUTE_COMPARE, mType);
	}

	/**
	 * @param allowAdornments Whether extras, such as "lbs." can be appended to the text.
	 * @return The numeric qualifier, as a {@link String}.
	 */
	public abstract String getQualifierAsString(boolean allowAdornments);

	/** @return The type of comparison to make. */
	public String getType() {
		return mType;
	}

	/**
	 * @param type The type of comparison to make. Must be one of {@link #IS}, {@link #AT_LEAST},
	 *            or {@link #NO_MORE_THAN}.
	 */
	public void setType(String type) {
		if (type == IS || type == AT_LEAST || type == NO_MORE_THAN) {
			mType = type;
		} else if (IS.equals(type)) {
			mType = IS;
		} else if (AT_LEAST.equals(type)) {
			mType = AT_LEAST;
		} else if (NO_MORE_THAN.equals(type)) {
			mType = NO_MORE_THAN;
		} else {
			mType = AT_LEAST;
		}
	}

	@Override public String toString() {
		return toString(Msgs.IS_PREFIX);
	}

	/**
	 * @param prefix A prefix to place before the description.
	 * @return A formatted description of this object.
	 */
	public String toString(String prefix) {
		if (IS == mType) {
			return MessageFormat.format(Msgs.IS_DESCRIPTION, prefix, getQualifierAsString(true));
		}
		if (AT_LEAST == mType) {
			return MessageFormat.format(Msgs.AT_LEAST, prefix, getQualifierAsString(true));
		}
		if (NO_MORE_THAN == mType) {
			return MessageFormat.format(Msgs.NO_MORE_THAN, prefix, getQualifierAsString(true));
		}
		return ""; //$NON-NLS-1$
	}
}
