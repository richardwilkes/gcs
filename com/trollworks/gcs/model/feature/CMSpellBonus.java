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
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/** A spell bonus. */
public class CMSpellBonus extends CMBonus {
	/** The XML tag. */
	public static final String	TAG_ROOT				= "spell_bonus";	//$NON-NLS-1$
	private static final String	TAG_COLLEGE_NAME		= "college_name";	//$NON-NLS-1$
	private static final String	TAG_SPELL_NAME			= "spell_name";	//$NON-NLS-1$
	/** The XML attribute name for the "all colleges" flag. */
	public static final String	ATTRIBUTE_ALL_COLLEGES	= "all_colleges";	//$NON-NLS-1$
	private boolean				mAllColleges;
	private boolean				mMatchCollegeName;
	private CMStringCriteria	mNameCriteria;

	/** Creates a new spell bonus. */
	public CMSpellBonus() {
		super(1);
		mAllColleges = true;
		mMatchCollegeName = true;
		mNameCriteria = new CMStringCriteria(CMStringCompareType.IS, ""); //$NON-NLS-1$
	}

	/**
	 * Loads a {@link CMSpellBonus}.
	 * 
	 * @param reader The XML reader to use.
	 * @throws IOException
	 */
	public CMSpellBonus(TKXMLReader reader) throws IOException {
		this();
		mAllColleges = reader.isAttributeSet(ATTRIBUTE_ALL_COLLEGES);
		mMatchCollegeName = true;
		load(reader);
	}

	/**
	 * Creates a clone of the specified bonus.
	 * 
	 * @param other The bonus to clone.
	 */
	public CMSpellBonus(CMSpellBonus other) {
		super(other);
		mAllColleges = other.mAllColleges;
		mMatchCollegeName = other.mMatchCollegeName;
		mNameCriteria = new CMStringCriteria(other.mNameCriteria);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMSpellBonus && super.equals(obj)) {
			CMSpellBonus other = (CMSpellBonus) obj;

			return mAllColleges == other.mAllColleges && mMatchCollegeName == other.mMatchCollegeName && mNameCriteria.equals(other.mNameCriteria);
		}
		return false;
	}

	public CMFeature cloneFeature() {
		return new CMSpellBonus(this);
	}

	public String getXMLTag() {
		return TAG_ROOT;
	}

	public String getKey() {
		StringBuffer buffer = new StringBuffer();

		if (mAllColleges) {
			buffer.append(CMSpell.ID_COLLEGE);
		} else {
			buffer.append(mMatchCollegeName ? CMSpell.ID_COLLEGE : CMSpell.ID_NAME);
			if (mNameCriteria.getType() == CMStringCompareType.IS) {
				buffer.append('/');
				buffer.append(mNameCriteria.getQualifier());
			} else {
				buffer.append("*"); //$NON-NLS-1$
			}
		}
		return buffer.toString();
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		String name = reader.getName();

		if (TAG_COLLEGE_NAME.equals(name)) {
			mNameCriteria.load(reader);
			mMatchCollegeName = true;
		} else if (TAG_SPELL_NAME.equals(name)) {
			mNameCriteria.load(reader);
			mMatchCollegeName = false;
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
		out.startTag(TAG_ROOT);
		if (mAllColleges) {
			out.writeAttribute(ATTRIBUTE_ALL_COLLEGES, mAllColleges);
		}
		out.finishTagEOL();
		if (!mAllColleges) {
			mNameCriteria.save(out, mMatchCollegeName ? TAG_COLLEGE_NAME : TAG_SPELL_NAME);
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

	/** @return Whether the bonus matches against the college name or the spell name. */
	public boolean matchesCollegeName() {
		return mMatchCollegeName;
	}

	/** @param college Whether the bonus matches against the college name or the spell name. */
	public void matchesCollegeName(boolean college) {
		mMatchCollegeName = college;
	}

	/** @return The college/spell name criteria. */
	public CMStringCriteria getNameCriteria() {
		return mNameCriteria;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		if (!mAllColleges) {
			CMRow.extractNameables(set, mNameCriteria.getQualifier());
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		if (!mAllColleges) {
			mNameCriteria.setQualifier(CMRow.nameNameables(map, mNameCriteria.getQualifier()));
		}
	}
}
