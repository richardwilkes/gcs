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

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** A skill bonus. */
public class CMSkillBonus extends CMBonus {
	/** The XML tag. */
	public static final String	TAG_ROOT			= "skill_bonus";	//$NON-NLS-1$
	private static final String	TAG_NAME			= "name";			//$NON-NLS-1$
	private static final String	TAG_SPECIALIZATION	= "specialization"; //$NON-NLS-1$
	private static final String	EMPTY				= "";				//$NON-NLS-1$
	private CMStringCriteria	mNameCriteria;
	private CMStringCriteria	mSpecializationCriteria;

	/** Creates a new skill bonus. */
	public CMSkillBonus() {
		super(1);
		mNameCriteria = new CMStringCriteria(CMStringCompareType.IS, EMPTY);
		mSpecializationCriteria = new CMStringCriteria(CMStringCompareType.IS_ANYTHING, EMPTY);
	}

	/**
	 * Loads a {@link CMSkillBonus}.
	 * 
	 * @param reader The XML reader to use.
	 * @throws IOException
	 */
	public CMSkillBonus(TKXMLReader reader) throws IOException {
		this();
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public CMSkillBonus(CMSkillBonus other) {
		super(other);
		mNameCriteria = new CMStringCriteria(other.mNameCriteria);
		mSpecializationCriteria = new CMStringCriteria(other.mSpecializationCriteria);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMSkillBonus && super.equals(obj)) {
			CMSkillBonus other = (CMSkillBonus) obj;

			return mNameCriteria.equals(other.mNameCriteria) && mSpecializationCriteria.equals(other.mSpecializationCriteria);
		}
		return false;
	}

	public CMFeature cloneFeature() {
		return new CMSkillBonus(this);
	}

	public String getXMLTag() {
		return TAG_ROOT;
	}

	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		buffer.append(CMSkill.ID_NAME);
		if (mNameCriteria.getType() == CMStringCompareType.IS && mSpecializationCriteria.getType() == CMStringCompareType.IS_ANYTHING) {
			buffer.append('/');
			buffer.append(mNameCriteria.getQualifier());
		} else {
			buffer.append("*"); //$NON-NLS-1$
		}
		return buffer.toString();
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_NAME.equals(reader.getName())) {
			mNameCriteria.load(reader);
		} else if (TAG_SPECIALIZATION.equals(reader.getName())) {
			mSpecializationCriteria.load(reader);
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
		mNameCriteria.save(out, TAG_NAME);
		mSpecializationCriteria.save(out, TAG_SPECIALIZATION);
		saveBase(out);
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The name criteria. */
	public CMStringCriteria getNameCriteria() {
		return mNameCriteria;
	}

	/** @return The name criteria. */
	public CMStringCriteria getSpecializationCriteria() {
		return mSpecializationCriteria;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		CMRow.extractNameables(set, mNameCriteria.getQualifier());
		CMRow.extractNameables(set, mSpecializationCriteria.getQualifier());
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		mNameCriteria.setQualifier(CMRow.nameNameables(map, mNameCriteria.getQualifier()));
		mSpecializationCriteria.setQualifier(CMRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
	}
}
