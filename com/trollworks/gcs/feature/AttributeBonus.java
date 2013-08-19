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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.ttk.collections.Enums;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;

/** An attribute bonus. */
public class AttributeBonus extends Bonus {
	/** The XML tag. */
	public static final String			TAG_ROOT				= "attribute_bonus";	//$NON-NLS-1$
	private static final String			TAG_ATTRIBUTE			= "attribute";			//$NON-NLS-1$
	private static final String			ATTRIBUTE_LIMITATION	= "limitation";		//$NON-NLS-1$
	private BonusAttributeType			mAttribute;
	private AttributeBonusLimitation	mLimitation;

	/** Creates a new attribute bonus. */
	public AttributeBonus() {
		super(1);
		mAttribute = BonusAttributeType.ST;
		mLimitation = AttributeBonusLimitation.NONE;
	}

	/**
	 * Loads a {@link AttributeBonus}.
	 * 
	 * @param reader The XML reader to use.
	 */
	public AttributeBonus(XMLReader reader) throws IOException {
		this();
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public AttributeBonus(AttributeBonus other) {
		super(other);
		mAttribute = other.mAttribute;
		mLimitation = other.mLimitation;
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof AttributeBonus && super.equals(obj)) {
			AttributeBonus ab = (AttributeBonus) obj;
			return mAttribute == ab.mAttribute && mLimitation == ab.mLimitation;
		}
		return false;
	}

	@Override
	public Feature cloneFeature() {
		return new AttributeBonus(this);
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(GURPSCharacter.ATTRIBUTES_PREFIX);
		buffer.append(mAttribute.name());
		if (mLimitation != AttributeBonusLimitation.NONE) {
			buffer.append(mLimitation.name());
		}
		return buffer.toString();
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		if (TAG_ATTRIBUTE.equals(reader.getName())) {
			setLimitation(Enums.extract(reader.getAttribute(ATTRIBUTE_LIMITATION), AttributeBonusLimitation.values(), AttributeBonusLimitation.NONE));
			setAttribute(Enums.extract(reader.readText(), BonusAttributeType.values(), BonusAttributeType.ST));
		} else {
			super.loadSelf(reader);
		}
	}

	/**
	 * Saves the bonus.
	 * 
	 * @param out The XML writer to use.
	 */
	@Override
	public void save(XMLWriter out) {
		out.startSimpleTagEOL(TAG_ROOT);
		out.startTag(TAG_ATTRIBUTE);
		if (mLimitation != AttributeBonusLimitation.NONE) {
			out.writeAttribute(ATTRIBUTE_LIMITATION, mLimitation.name().toLowerCase());
		}
		out.finishTag();
		out.writeEncodedData(mAttribute.name().toLowerCase());
		out.endTagEOL(TAG_ATTRIBUTE, false);
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The attribute this bonus applies to. */
	public BonusAttributeType getAttribute() {
		return mAttribute;
	}

	/** @param attribute The attribute. */
	public void setAttribute(BonusAttributeType attribute) {
		mAttribute = attribute;
		getAmount().setIntegerOnly(mAttribute.isIntegerOnly());
	}

	/** @return The limitation of this bonus. */
	public AttributeBonusLimitation getLimitation() {
		return mLimitation;
	}

	/** @param limitation The limitation. */
	public void setLimitation(AttributeBonusLimitation limitation) {
		mLimitation = limitation;
	}
}
