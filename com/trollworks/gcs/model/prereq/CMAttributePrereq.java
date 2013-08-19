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
import com.trollworks.gcs.model.criteria.CMNumericCompareType;
import com.trollworks.gcs.model.feature.CMBonusAttributeType;
import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;

/** A Attribute prerequisite. */
public class CMAttributePrereq extends CMHasPrereq {
	/** The possible {@link CMBonusAttributeType}s that can be affected. */
	public static final CMBonusAttributeType[]	TYPES					= { CMBonusAttributeType.ST, CMBonusAttributeType.DX, CMBonusAttributeType.IQ, CMBonusAttributeType.HT, CMBonusAttributeType.WILL };
	/** The XML tag for this class. */
	public static final String					TAG_ROOT				= "attribute_prereq";																													//$NON-NLS-1$
	private static final String					ATTRIBUTE_WHICH			= "which";																																//$NON-NLS-1$
	private static final String					ATTRIBUTE_COMBINED_WITH	= "combined_with";																														//$NON-NLS-1$
	private static final String					ATTRIBUTE_COMPARE		= "compare";																															//$NON-NLS-1$
	private CMBonusAttributeType				mWhich;
	private CMBonusAttributeType				mCombinedWith;
	private CMIntegerCriteria					mValueCompare;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMAttributePrereq(CMPrereqList parent) {
		super(parent);
		mValueCompare = new CMIntegerCriteria(CMNumericCompareType.AT_LEAST, 10);
		setWhich(CMBonusAttributeType.IQ);
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
		setWhich((CMBonusAttributeType) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_WHICH), TYPES, CMBonusAttributeType.ST));
		setCombinedWith((CMBonusAttributeType) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_COMBINED_WITH), TYPES));
		mValueCompare.setType((CMNumericCompareType) TKEnumExtractor.extract(reader.getAttribute(ATTRIBUTE_COMPARE), CMNumericCompareType.values(), CMNumericCompareType.AT_LEAST));
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

			return mWhich == other.mWhich && mCombinedWith == other.mCombinedWith && mValueCompare.equals(other.mValueCompare);
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
		out.writeAttribute(ATTRIBUTE_WHICH, mWhich.name().toLowerCase());
		if (mCombinedWith != null) {
			out.writeAttribute(ATTRIBUTE_COMBINED_WITH, mCombinedWith.name().toLowerCase());
		}
		out.writeAttribute(ATTRIBUTE_COMPARE, mValueCompare.getType().name().toLowerCase());
		out.finishTag();
		out.writeEncodedData(Integer.toString(mValueCompare.getQualifier()));
		out.endTagEOL(TAG_ROOT, false);
	}

	/** @return The type of comparison to make. */
	public CMBonusAttributeType getWhich() {
		return mWhich;
	}

	/** @param which The type of comparison to make. */
	public void setWhich(CMBonusAttributeType which) {
		mWhich = which;
	}

	/** @return The type of comparison to make. */
	public CMBonusAttributeType getCombinedWith() {
		return mCombinedWith;
	}

	/** @param which The type of comparison to make. */
	public void setCombinedWith(CMBonusAttributeType which) {
		mCombinedWith = which;
	}

	private int getAttributeValue(CMCharacter character, CMBonusAttributeType attribute) {
		if (attribute == null) {
			return 0;
		}
		switch (attribute) {
			case ST:
				return character.getStrength();
			case DX:
				return character.getDexterity();
			case IQ:
				return character.getIntelligence();
			case HT:
				return character.getHealth();
			case WILL:
				return character.getWill();
			default:
				return 0;
		}
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
			builder.append(MessageFormat.format(Msgs.DESCRIPTION, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mCombinedWith == null ? mWhich.getPresentationName() : MessageFormat.format(Msgs.COMBINED, mWhich.getPresentationName(), mCombinedWith.getPresentationName()), mValueCompare.toString()));
		}
		return satisfied;
	}
}
