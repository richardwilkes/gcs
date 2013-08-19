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

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.criteria.CMIntegerCriteria;
import com.trollworks.gcs.model.criteria.CMNumericCompareType;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/**
 * An abstract prerequisite class for comparison of name and level and whether or not the specific
 * item is present.
 */
public abstract class CMNameLevelPrereq extends CMHasPrereq {
	private static final String	TAG_NAME	= "name";	//$NON-NLS-1$
	private static final String	TAG_LEVEL	= "level";	//$NON-NLS-1$
	private String				mTag;
	private CMStringCriteria	mNameCriteria;
	private CMIntegerCriteria	mLevelCriteria;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param tag The tag for this prerequisite.
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMNameLevelPrereq(String tag, CMPrereqList parent) {
		super(parent);
		mTag = tag;
		mNameCriteria = new CMStringCriteria(CMStringCompareType.IS, ""); //$NON-NLS-1$
		mLevelCriteria = new CMIntegerCriteria(CMNumericCompareType.AT_LEAST, 0);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMNameLevelPrereq(CMPrereqList parent, TKXMLReader reader) throws IOException {
		this(reader.getName(), parent);
		initializeForLoad();
		String marker = reader.getMarker();
		loadHasAttribute(reader);

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				loadSelf(reader);
			}
		} while (reader.withinMarker(marker));
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected CMNameLevelPrereq(CMPrereqList parent, CMNameLevelPrereq prereq) {
		super(parent, prereq);
		mTag = prereq.mTag;
		mNameCriteria = new CMStringCriteria(prereq.mNameCriteria);
		mLevelCriteria = new CMIntegerCriteria(prereq.mLevelCriteria);
	}

	/** Called so that sub-classes can initialize themselves prior to loading. */
	protected void initializeForLoad() {
		// Does nothing
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMNameLevelPrereq && super.equals(obj)) {
			CMNameLevelPrereq other = (CMNameLevelPrereq) obj;

			return mTag.equals(other.mTag) && mNameCriteria.equals(other.mNameCriteria) && mLevelCriteria.equals(other.mLevelCriteria);
		}
		return false;
	}

	@Override public void save(TKXMLWriter out) {
		out.startTag(mTag);
		saveHasAttribute(out);
		out.finishTagEOL();
		mNameCriteria.save(out, TAG_NAME);
		saveSelf(out);
		if (mLevelCriteria.getType() != CMNumericCompareType.AT_LEAST || mLevelCriteria.getQualifier() != 0) {
			mLevelCriteria.save(out, TAG_LEVEL);
		}
		out.endTagEOL(mTag, true);
	}

	/**
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	protected void loadSelf(TKXMLReader reader) throws IOException {
		String name = reader.getName();

		if (TAG_NAME.equals(name)) {
			mNameCriteria.load(reader);
		} else if (TAG_LEVEL.equals(name)) {
			mLevelCriteria.load(reader);
		} else {
			reader.skipTag(name);
		}
	}

	/**
	 * Called so that sub-classes can save extra data.
	 * 
	 * @param out The XML writer to use.
	 */
	protected void saveSelf(@SuppressWarnings("unused") TKXMLWriter out) {
		// Does nothing
	}

	/** @return The name comparison object. */
	public CMStringCriteria getNameCriteria() {
		return mNameCriteria;
	}

	/** @return The level comparison object. */
	public CMIntegerCriteria getLevelCriteria() {
		return mLevelCriteria;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		CMRow.extractNameables(set, mNameCriteria.getQualifier());
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		mNameCriteria.setQualifier(CMRow.nameNameables(map, mNameCriteria.getQualifier()));
	}
}
