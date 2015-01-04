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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.text.Enums;

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
	public int hashCode() {
		return super.hashCode();
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
			out.writeAttribute(ATTRIBUTE_LIMITATION, Enums.toId(mLimitation));
		}
		out.finishTag();
		out.writeEncodedData(Enums.toId(mAttribute));
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
