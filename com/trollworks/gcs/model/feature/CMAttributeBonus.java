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
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;

/** An attribute bonus. */
public class CMAttributeBonus extends CMBonus {
	/** The XML tag. */
	public static final String	TAG_ROOT				= "attribute_bonus";	//$NON-NLS-1$
	private static final String	TAG_ATTRIBUTE			= "attribute";			//$NON-NLS-1$
	private static final String	ATTRIBUTE_LIMITATION	= "limitation";		//$NON-NLS-1$
	/** The ST attribute. */
	public static final String	ST						= "ST";				//$NON-NLS-1$
	/** The DX attribute. */
	public static final String	DX						= "DX";				//$NON-NLS-1$
	/** The IQ attribute. */
	public static final String	IQ						= "IQ";				//$NON-NLS-1$
	/** The HT attribute. */
	public static final String	HT						= "HT";				//$NON-NLS-1$
	/** The Will attribute. */
	public static final String	WILL					= "will";				//$NON-NLS-1$
	/** The Perception attribute. */
	public static final String	PER						= "perception";		//$NON-NLS-1$
	/** The Vision attribute. */
	public static final String	VISION					= "vision";			//$NON-NLS-1$
	/** The Hearing attribute. */
	public static final String	HEARING					= "hearing";			//$NON-NLS-1$
	/** The TasteSmell attribute. */
	public static final String	TASTE_SMELL				= "taste,smell";		//$NON-NLS-1$
	/** The Touch attribute. */
	public static final String	TOUCH					= "touch";				//$NON-NLS-1$
	/** The Dodge attribute. */
	public static final String	DODGE					= "dodge";				//$NON-NLS-1$
	/** The Dodge attribute. */
	public static final String	PARRY					= "parry";				//$NON-NLS-1$
	/** The Dodge attribute. */
	public static final String	BLOCK					= "block";				//$NON-NLS-1$
	/** The Speed attribute. */
	public static final String	SPEED					= "speed";				//$NON-NLS-1$
	/** The Move attribute. */
	public static final String	MOVE					= "move";				//$NON-NLS-1$
	/** The FP attribute. */
	public static final String	FP						= "FP";				//$NON-NLS-1$
	/** The HP attribute. */
	public static final String	HP						= "HP";				//$NON-NLS-1$
	/** The size modifier attribute. */
	public static final String	SM						= "SM";				//$NON-NLS-1$
	/** The "for striking only" attribute. */
	public static final String	STRIKING_ONLY			= "striking_only";		//$NON-NLS-1$
	/** The "for lifting only" attribute. */
	public static final String	LIFTING_ONLY			= "lifting_only";		//$NON-NLS-1$
	private String				mAttribute;
	private String				mLimitation;

	/** Creates a new attribute bonus. */
	public CMAttributeBonus() {
		super(1);
		mAttribute = ST;
		mLimitation = null;
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

			return mAttribute.equals(other.mAttribute) && (mLimitation == null ? other.mLimitation == null : mLimitation.equals(other.mLimitation));
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
		buffer.append(mAttribute);
		if (mLimitation != null) {
			buffer.append(mLimitation);
		}
		return buffer.toString();
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_ATTRIBUTE.equals(reader.getName())) {
			setLimitation(reader.getAttribute(ATTRIBUTE_LIMITATION));
			setAttribute(reader.readText());
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
		if (mLimitation != null) {
			out.writeAttribute(ATTRIBUTE_LIMITATION, mLimitation);
		}
		out.finishTag();
		out.writeEncodedData(mAttribute);
		out.endTagEOL(TAG_ATTRIBUTE, false);
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The attribute this bonus applies to. */
	public String getAttribute() {
		return mAttribute;
	}

	/** @param attribute The attribute. */
	public void setAttribute(String attribute) {
		boolean integerOnly = true;

		if (ST.equals(attribute)) {
			mAttribute = ST;
		} else if (DX.equals(attribute)) {
			mAttribute = DX;
		} else if (IQ.equals(attribute)) {
			mAttribute = IQ;
		} else if (HT.equals(attribute)) {
			mAttribute = HT;
		} else if (WILL.equals(attribute)) {
			mAttribute = WILL;
		} else if (PER.equals(attribute)) {
			mAttribute = PER;
		} else if (VISION.equals(attribute)) {
			mAttribute = VISION;
		} else if (HEARING.equals(attribute)) {
			mAttribute = HEARING;
		} else if (TASTE_SMELL.equals(attribute)) {
			mAttribute = TASTE_SMELL;
		} else if (TOUCH.equals(attribute)) {
			mAttribute = TOUCH;
		} else if (DODGE.equals(attribute)) {
			mAttribute = DODGE;
		} else if (PARRY.equals(attribute)) {
			mAttribute = PARRY;
		} else if (BLOCK.equals(attribute)) {
			mAttribute = BLOCK;
		} else if (SPEED.equals(attribute)) {
			mAttribute = SPEED;
			integerOnly = false;
		} else if (MOVE.equals(attribute)) {
			mAttribute = MOVE;
		} else if (FP.equals(attribute)) {
			mAttribute = FP;
		} else if (HP.equals(attribute)) {
			mAttribute = HP;
		} else if (SM.equals(attribute)) {
			mAttribute = SM;
		} else {
			mAttribute = ST;
		}

		getAmount().setIntegerOnly(integerOnly);
	}

	/**
	 * @return The limitation of this bonus. <code>null</code> will be returned if there is no
	 *         limitation.
	 */
	public String getLimitation() {
		return mLimitation;
	}

	/** @param limitation The limitation. */
	public void setLimitation(String limitation) {
		if (limitation == null || LIFTING_ONLY == limitation || STRIKING_ONLY == limitation) {
			mLimitation = limitation;
		} else if (LIFTING_ONLY.equals(limitation)) {
			mLimitation = LIFTING_ONLY;
		} else if (STRIKING_ONLY.equals(limitation)) {
			mLimitation = STRIKING_ONLY;
		} else {
			mLimitation = null;
		}
	}
}
