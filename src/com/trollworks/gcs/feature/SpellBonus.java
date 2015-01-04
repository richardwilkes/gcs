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

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** A spell bonus. */
public class SpellBonus extends Bonus {
	/** The XML tag. */
	public static final String	TAG_ROOT				= "spell_bonus";		//$NON-NLS-1$
	/** Matches against the college name. */
	public static final String	TAG_COLLEGE_NAME		= "college_name";		//$NON-NLS-1$
	/** Matches against the power source name. */
	public static final String	TAG_POWER_SOURCE_NAME	= "power_source_name";	//$NON-NLS-1$
	/** Matches against the spell name. */
	public static final String	TAG_SPELL_NAME			= "spell_name";		//$NON-NLS-1$
	/** The XML attribute name for the "all colleges" flag. */
	public static final String	ATTRIBUTE_ALL_COLLEGES	= "all_colleges";		//$NON-NLS-1$
	private boolean				mAllColleges;
	private String				mMatchType;
	private StringCriteria		mNameCriteria;

	/** Creates a new spell bonus. */
	public SpellBonus() {
		super(1);
		mAllColleges = true;
		mMatchType = TAG_COLLEGE_NAME;
		mNameCriteria = new StringCriteria(StringCompareType.IS, ""); //$NON-NLS-1$
	}

	/**
	 * Loads a {@link SpellBonus}.
	 * 
	 * @param reader The XML reader to use.
	 */
	public SpellBonus(XMLReader reader) throws IOException {
		this();
		mAllColleges = reader.isAttributeSet(ATTRIBUTE_ALL_COLLEGES);
		mMatchType = TAG_COLLEGE_NAME;
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public SpellBonus(SpellBonus other) {
		super(other);
		mAllColleges = other.mAllColleges;
		mMatchType = other.mMatchType;
		mNameCriteria = new StringCriteria(other.mNameCriteria);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof SpellBonus && super.equals(obj)) {
			SpellBonus sb = (SpellBonus) obj;
			return mAllColleges == sb.mAllColleges && mMatchType == sb.mMatchType && mNameCriteria.equals(sb.mNameCriteria);
		}
		return false;
	}

	@Override
	public Feature cloneFeature() {
		return new SpellBonus(this);
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		if (mAllColleges) {
			buffer.append(Spell.ID_COLLEGE);
		} else {
			if (mMatchType == TAG_COLLEGE_NAME) {
				buffer.append(Spell.ID_COLLEGE);
			} else if (mMatchType == TAG_POWER_SOURCE_NAME) {
				buffer.append(Spell.ID_POWER_SOURCE);
			} else {
				buffer.append(Spell.ID_NAME);
			}
			if (mNameCriteria.getType() == StringCompareType.IS) {
				buffer.append('/');
				buffer.append(mNameCriteria.getQualifier());
			} else {
				buffer.append("*"); //$NON-NLS-1$
			}
		}
		return buffer.toString();
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		String name = reader.getName();
		if (TAG_COLLEGE_NAME.equals(name)) {
			mMatchType = TAG_COLLEGE_NAME;
			mNameCriteria.load(reader);
		} else if (TAG_POWER_SOURCE_NAME.equals(name)) {
			mMatchType = TAG_POWER_SOURCE_NAME;
			mNameCriteria.load(reader);
		} else if (TAG_SPELL_NAME.equals(name)) {
			mMatchType = TAG_SPELL_NAME;
			mNameCriteria.load(reader);
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
		out.startTag(TAG_ROOT);
		if (mAllColleges) {
			out.writeAttribute(ATTRIBUTE_ALL_COLLEGES, mAllColleges);
		}
		out.finishTagEOL();
		if (!mAllColleges) {
			mNameCriteria.save(out, mMatchType);
		}
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return Whether the bonus applies to all colleges. */
	public boolean allColleges() {
		return mAllColleges;
	}

	/** @param all Whether the bonus applies to all colleges. */
	public void allColleges(boolean all) {
		mAllColleges = all;
	}

	/**
	 * @return The match type. One of {@link #TAG_COLLEGE_NAME}, {@link #TAG_POWER_SOURCE_NAME}, or
	 *         {@link #TAG_SPELL_NAME}.
	 */
	public String getMatchType() {
		return mMatchType;
	}

	public void setMatchType(String type) {
		if (TAG_COLLEGE_NAME.equals(type)) {
			mMatchType = TAG_COLLEGE_NAME;
		} else if (TAG_POWER_SOURCE_NAME.equals(type)) {
			mMatchType = TAG_POWER_SOURCE_NAME;
		} else {
			mMatchType = TAG_SPELL_NAME;
		}
	}

	/** @return The college/spell name criteria. */
	public StringCriteria getNameCriteria() {
		return mNameCriteria;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		if (!mAllColleges) {
			ListRow.extractNameables(set, mNameCriteria.getQualifier());
		}
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		if (!mAllColleges) {
			mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
		}
	}
}
