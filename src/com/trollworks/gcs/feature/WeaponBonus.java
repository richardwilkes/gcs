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

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** A weapon bonus. */
public class WeaponBonus extends Bonus {
	/** The XML tag. */
	public static final String	TAG_ROOT			= "weapon_bonus";	//$NON-NLS-1$
	private static final String	TAG_NAME			= "name";			//$NON-NLS-1$
	private static final String	TAG_SPECIALIZATION	= "specialization"; //$NON-NLS-1$
	private static final String	TAG_LEVEL			= "level";			//$NON-NLS-1$
	private static final String	EMPTY				= "";				//$NON-NLS-1$
	private StringCriteria		mNameCriteria;
	private StringCriteria		mSpecializationCriteria;
	private IntegerCriteria		mLevelCriteria;

	/** Creates a new skill bonus. */
	public WeaponBonus() {
		super(1);
		mNameCriteria = new StringCriteria(StringCompareType.IS, EMPTY);
		mSpecializationCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, EMPTY);
		mLevelCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 0);
	}

	/**
	 * Loads a {@link WeaponBonus}.
	 * 
	 * @param reader The XML reader to use.
	 */
	public WeaponBonus(XMLReader reader) throws IOException {
		this();
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public WeaponBonus(WeaponBonus other) {
		super(other);
		mNameCriteria = new StringCriteria(other.mNameCriteria);
		mSpecializationCriteria = new StringCriteria(other.mSpecializationCriteria);
		mLevelCriteria = new IntegerCriteria(other.mLevelCriteria);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof WeaponBonus && super.equals(obj)) {
			WeaponBonus wb = (WeaponBonus) obj;
			return mNameCriteria.equals(wb.mNameCriteria) && mSpecializationCriteria.equals(wb.mSpecializationCriteria) && mLevelCriteria.equals(wb.mLevelCriteria);
		}
		return false;
	}

	@Override
	public Feature cloneFeature() {
		return new WeaponBonus(this);
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(Skill.ID_NAME);
		if (mNameCriteria.getType() == StringCompareType.IS && mSpecializationCriteria.getType() == StringCompareType.IS_ANYTHING) {
			buffer.append('/');
			buffer.append(mNameCriteria.getQualifier());
		} else {
			buffer.append("*"); //$NON-NLS-1$
		}
		return buffer.toString();
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		if (TAG_NAME.equals(reader.getName())) {
			mNameCriteria.load(reader);
		} else if (TAG_SPECIALIZATION.equals(reader.getName())) {
			mSpecializationCriteria.load(reader);
		} else if (TAG_LEVEL.equals(reader.getName())) {
			mLevelCriteria.load(reader);
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
		mNameCriteria.save(out, TAG_NAME);
		mSpecializationCriteria.save(out, TAG_SPECIALIZATION);
		mLevelCriteria.save(out, TAG_LEVEL);
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The name criteria. */
	public StringCriteria getNameCriteria() {
		return mNameCriteria;
	}

	/** @return The name criteria. */
	public StringCriteria getSpecializationCriteria() {
		return mSpecializationCriteria;
	}

	/** @return The level criteria. */
	public IntegerCriteria getLevelCriteria() {
		return mLevelCriteria;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		ListRow.extractNameables(set, mNameCriteria.getQualifier());
		ListRow.extractNameables(set, mSpecializationCriteria.getQualifier());
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
		mSpecializationCriteria.setQualifier(ListRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
	}
}
