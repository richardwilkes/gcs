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

package com.trollworks.gcs.model.prereq;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMIntegerCriteria;
import com.trollworks.gcs.model.criteria.CMNumericCriteria;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;

/** A Attribute prerequisite. */
public class CMAttributePrereq extends CMHasPrereq {
	/** The XML tag for this class. */
	public static final String	TAG_ROOT				= "attribute_prereq";	//$NON-NLS-1$
	private static final String	ATTRIBUTE_WHICH			= "which";				//$NON-NLS-1$
	private static final String	ATTRIBUTE_COMBINED_WITH	= "combined_with";		//$NON-NLS-1$
	private static final String	ATTRIBUTE_COMPARE		= "compare";			//$NON-NLS-1$
	/** The constant for strength. */
	public static final String	ST						= "ST";				//$NON-NLS-1$
	/** The constant for dexterity. */
	public static final String	DX						= "DX";				//$NON-NLS-1$
	/** The constant for intelligence. */
	public static final String	IQ						= "IQ";				//$NON-NLS-1$
	/** The constant for health. */
	public static final String	HT						= "HT";				//$NON-NLS-1$
	/** The constant for will. */
	public static final String	WILL					= "Will";				//$NON-NLS-1$
	private String				mWhich;
	private String				mCombinedWith;
	private CMIntegerCriteria	mValueCompare;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMAttributePrereq(CMPrereqList parent) {
		super(parent);
		mValueCompare = new CMIntegerCriteria(CMNumericCriteria.AT_LEAST, 10);
		setWhich(IQ);
		setCombinedWith(null);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMAttributePrereq(CMPrereqList parent, TKXMLReader reader) throws IOException {
		this(parent);
		loadHasAttribute(reader);
		setWhich(reader.getAttribute(ATTRIBUTE_WHICH));
		setCombinedWith(reader.getAttribute(ATTRIBUTE_COMBINED_WITH));
		mValueCompare.setType(reader.getAttribute(ATTRIBUTE_COMPARE));
		mValueCompare.setQualifier(reader.readInteger(10));
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected CMAttributePrereq(CMPrereqList parent, CMAttributePrereq prereq) {
		super(parent, prereq);
		mWhich = prereq.mWhich;
		mCombinedWith = prereq.mCombinedWith;
		mValueCompare = new CMIntegerCriteria(prereq.mValueCompare);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMAttributePrereq && super.equals(obj)) {
			CMAttributePrereq other = (CMAttributePrereq) obj;

			return mWhich.equals(other.mWhich) && (mCombinedWith == null ? other.mCombinedWith == null : mCombinedWith.equals(other.mCombinedWith)) && mValueCompare.equals(other.mValueCompare);
		}
		return false;
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public CMPrereq clone(CMPrereqList parent) {
		return new CMAttributePrereq(parent, this);
	}

	@Override public void save(TKXMLWriter out) {
		out.startTag(TAG_ROOT);
		saveHasAttribute(out);
		out.writeAttribute(ATTRIBUTE_WHICH, mWhich);
		if (mCombinedWith != null) {
			out.writeAttribute(ATTRIBUTE_COMBINED_WITH, mCombinedWith);
		}
		out.writeAttribute(ATTRIBUTE_COMPARE, mValueCompare.getType());
		out.finishTag();
		out.writeEncodedData(Integer.toString(mValueCompare.getQualifier()));
		out.endTagEOL(TAG_ROOT, false);
	}

	/** @return The type of comparison to make. */
	public String getWhich() {
		return mWhich;
	}

	/**
	 * @param which The type of comparison to make. Must be one of {@link #ST}, {@link #DX},
	 *            {@link #IQ}, {@link #HT}, or {@link #WILL}.
	 */
	public void setWhich(String which) {
		mWhich = getMatchingAttribute(which);
		if (mWhich == null) {
			mWhich = ST;
		}
	}

	/** @return The type of comparison to make. */
	public String getCombinedWith() {
		return mCombinedWith;
	}

	/**
	 * @param which The type of comparison to make. Must be one of {@link #ST}, {@link #DX},
	 *            {@link #IQ}, {@link #HT}, or {@link #WILL}.
	 */
	public void setCombinedWith(String which) {
		mCombinedWith = getMatchingAttribute(which);
	}

	private String getMatchingAttribute(String attribute) {
		if (attribute == ST || attribute == DX || attribute == IQ || attribute == HT || attribute == WILL) {
			return attribute;
		} else if (ST.equals(attribute)) {
			return ST;
		} else if (DX.equals(attribute)) {
			return DX;
		} else if (IQ.equals(attribute)) {
			return IQ;
		} else if (HT.equals(attribute)) {
			return HT;
		} else if (WILL.equals(attribute)) {
			return WILL;
		}
		return null;
	}

	private int getAttributeValue(CMCharacter character, String attribute) {
		if (attribute == ST) {
			return character.getStrength();
		}
		if (attribute == DX) {
			return character.getDexterity();
		}
		if (attribute == IQ) {
			return character.getIntelligence();
		}
		if (attribute == HT) {
			return character.getHealth();
		}
		if (attribute == WILL) {
			return character.getWill();
		}
		return 0;
	}

	/** @return The value comparison object. */
	public CMIntegerCriteria getValueCompare() {
		return mValueCompare;
	}

	@Override public boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = mValueCompare.matches(getAttributeValue(character, mWhich) + getAttributeValue(character, mCombinedWith));

		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(Msgs.DESCRIPTION, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mCombinedWith == null ? mWhich : MessageFormat.format(Msgs.COMBINED, mWhich, mCombinedWith), mValueCompare.toString()));
		}
		return satisfied;
	}
}
