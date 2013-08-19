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

package com.trollworks.gcs.model.feature;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;

/** An attribute bonus. */
public class CMAttributeBonus extends CMBonus {
	/** The XML tag. */
	public static final String			TAG_ROOT				= "attribute_bonus";	//$NON-NLS-1$
	private static final String			TAG_ATTRIBUTE			= "attribute";			//$NON-NLS-1$
	private static final String			ATTRIBUTE_LIMITATION	= "limitation";		//$NON-NLS-1$
	private CMBonusAttributeType		mAttribute;
	private CMAttributeBonusLimitation	mLimitation;

	/** Creates a new attribute bonus. */
	public CMAttributeBonus() {
		super(1);
		mAttribute = CMBonusAttributeType.ST;
		mLimitation = CMAttributeBonusLimitation.NONE;
	}

	/**
	 * Loads a {@link CMAttributeBonus}.
	 * 
	 * @param reader The XML reader to use.
	 * @throws IOException
	 */
	public CMAttributeBonus(TKXMLReader reader) throws IOException {
		this();
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public CMAttributeBonus(CMAttributeBonus other) {
		super(other);
		mAttribute = other.mAttribute;
		mLimitation = other.mLimitation;
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMAttributeBonus && super.equals(obj)) {
			CMAttributeBonus other = (CMAttributeBonus) obj;

			return mAttribute == other.mAttribute && mLimitation == other.mLimitation;
		}
		return false;
	}

	public CMFeature cloneFeature() {
		return new CMAttributeBonus(this);
	}

	public String getXMLTag() {
		return TAG_ROOT;
	}

	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(CMCharacter.ATTRIBUTES_PREFIX);
		buffer.append(mAttribute.name());
		if (mLimitation != CMAttributeBonusLimitation.NONE) {
			buffer.append(mLimitation.name());
		}
		return buffer.toString();
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_ATTRIBUTE.equals(reader.getName())) {
			setLimitation((CMAttributeBonusLimitation) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_LIMITATION), CMAttributeBonusLimitation.values(), CMAttributeBonusLimitation.NONE));
			setAttribute((CMBonusAttributeType) TKEnumExtractor.extract(reader.readText(), CMBonusAttributeType.values(), CMBonusAttributeType.ST));
		} else {
			super.loadSelf(reader);
		}
	}

	/**
	 * Saves the bonus.
	 * 
	 * @param out The XML writer to use.
	 */
	public void save(TKXMLWriter out) {
		out.startSimpleTagEOL(TAG_ROOT);
		out.startTag(TAG_ATTRIBUTE);
		if (mLimitation != CMAttributeBonusLimitation.NONE) {
			out.writeAttribute(ATTRIBUTE_LIMITATION, mLimitation.name().toLowerCase());
		}
		out.finishTag();
		out.writeEncodedData(mAttribute.name().toLowerCase());
		out.endTagEOL(TAG_ATTRIBUTE, false);
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The attribute this bonus applies to. */
	public CMBonusAttributeType getAttribute() {
		return mAttribute;
	}

	/** @param attribute The attribute. */
	public void setAttribute(CMBonusAttributeType attribute) {
		mAttribute = attribute;
		getAmount().setIntegerOnly(mAttribute.isIntegerOnly());
	}

	/** @return The limitation of this bonus. */
	public CMAttributeBonusLimitation getLimitation() {
		return mLimitation;
	}

	/** @param limitation The limitation. */
	public void setLimitation(CMAttributeBonusLimitation limitation) {
		mLimitation = limitation;
	}
}
