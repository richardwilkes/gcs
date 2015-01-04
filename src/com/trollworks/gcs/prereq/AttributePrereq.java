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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.feature.BonusAttributeType;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Enums;

import java.io.IOException;
import java.text.MessageFormat;

/** A Attribute prerequisite. */
public class AttributePrereq extends HasPrereq {
	@Localize("{0}{1} {2} which {3}")
	@Localize(locale = "de", value = "{0}{1} {2} die {3}")
	@Localize(locale = "ru", value = "{0}{1} {2} который {3}")
	private static String						DESCRIPTION;
	@Localize("{0}+{1}")
	@Localize(locale = "de", value = "{0}+{1}")
	private static String						COMBINED;

	static {
		Localization.initialize();
	}

	/** The possible {@link BonusAttributeType}s that can be affected. */
	public static final BonusAttributeType[]	TYPES					= { BonusAttributeType.ST, BonusAttributeType.DX, BonusAttributeType.IQ, BonusAttributeType.HT, BonusAttributeType.WILL, BonusAttributeType.PERCEPTION };
	/** The XML tag for this class. */
	public static final String					TAG_ROOT				= "attribute_prereq";																																		//$NON-NLS-1$
	private static final String					ATTRIBUTE_WHICH			= "which";																																					//$NON-NLS-1$
	private static final String					ATTRIBUTE_COMBINED_WITH	= "combined_with";																																			//$NON-NLS-1$
	private static final String					ATTRIBUTE_COMPARE		= "compare";																																				//$NON-NLS-1$
	private BonusAttributeType					mWhich;
	private BonusAttributeType					mCombinedWith;
	private IntegerCriteria						mValueCompare;

	/**
	 * Creates a new prerequisite.
	 *
	 * @param parent The owning prerequisite list, if any.
	 */
	public AttributePrereq(PrereqList parent) {
		super(parent);
		mValueCompare = new IntegerCriteria(NumericCompareType.AT_LEAST, 10);
		setWhich(BonusAttributeType.IQ);
		setCombinedWith(null);
	}

	/**
	 * Loads a prerequisite.
	 *
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 */
	public AttributePrereq(PrereqList parent, XMLReader reader) throws IOException {
		this(parent);
		loadHasAttribute(reader);
		setWhich(Enums.extract(reader.getAttribute(ATTRIBUTE_WHICH), TYPES, BonusAttributeType.ST));
		setCombinedWith(Enums.extract(reader.getAttribute(ATTRIBUTE_COMBINED_WITH), TYPES));
		mValueCompare.setType(Enums.extract(reader.getAttribute(ATTRIBUTE_COMPARE), NumericCompareType.values(), NumericCompareType.AT_LEAST));
		mValueCompare.setQualifier(reader.readInteger(10));
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 *
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected AttributePrereq(PrereqList parent, AttributePrereq prereq) {
		super(parent, prereq);
		mWhich = prereq.mWhich;
		mCombinedWith = prereq.mCombinedWith;
		mValueCompare = new IntegerCriteria(prereq.mValueCompare);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof AttributePrereq && super.equals(obj)) {
			AttributePrereq ap = (AttributePrereq) obj;
			return mWhich == ap.mWhich && mCombinedWith == ap.mCombinedWith && mValueCompare.equals(ap.mValueCompare);
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public Prereq clone(PrereqList parent) {
		return new AttributePrereq(parent, this);
	}

	@Override
	public void save(XMLWriter out) {
		out.startTag(TAG_ROOT);
		saveHasAttribute(out);
		out.writeAttribute(ATTRIBUTE_WHICH, Enums.toId(mWhich));
		if (mCombinedWith != null) {
			out.writeAttribute(ATTRIBUTE_COMBINED_WITH, Enums.toId(mCombinedWith));
		}
		out.writeAttribute(ATTRIBUTE_COMPARE, Enums.toId(mValueCompare.getType()));
		out.finishTag();
		out.writeEncodedData(Integer.toString(mValueCompare.getQualifier()));
		out.endTagEOL(TAG_ROOT, false);
	}

	/** @return The type of comparison to make. */
	public BonusAttributeType getWhich() {
		return mWhich;
	}

	/** @param which The type of comparison to make. */
	public void setWhich(BonusAttributeType which) {
		mWhich = which;
	}

	/** @return The type of comparison to make. */
	public BonusAttributeType getCombinedWith() {
		return mCombinedWith;
	}

	/** @param which The type of comparison to make. */
	public void setCombinedWith(BonusAttributeType which) {
		mCombinedWith = which;
	}

	private static int getAttributeValue(GURPSCharacter character, BonusAttributeType attribute) {
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
			case PERCEPTION:
				return character.getPerception();
			default:
				return 0;
		}
	}

	/** @return The value comparison object. */
	public IntegerCriteria getValueCompare() {
		return mValueCompare;
	}

	@Override
	public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = mValueCompare.matches(getAttributeValue(character, mWhich) + getAttributeValue(character, mCombinedWith));

		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(DESCRIPTION, prefix, has() ? HAS : DOES_NOT_HAVE, mCombinedWith == null ? mWhich.getPresentationName() : MessageFormat.format(COMBINED, mWhich.getPresentationName(), mCombinedWith.getPresentationName()), mValueCompare.toString()));
		}
		return satisfied;
	}
}
