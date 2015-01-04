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

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.io.xml.XMLNodeType;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;

import java.io.IOException;
import java.util.HashMap;
import java.util.HashSet;

/**
 * An abstract prerequisite class for comparison of name and level and whether or not the specific
 * item is present.
 */
public abstract class NameLevelPrereq extends HasPrereq {
	/** Provided for sub-classes. */
	private static final String	TAG_NAME	= "name";	//$NON-NLS-1$
	private static final String	TAG_LEVEL	= "level";	//$NON-NLS-1$
	private String				mTag;
	private StringCriteria		mNameCriteria;
	private IntegerCriteria		mLevelCriteria;

	/**
	 * Creates a new prerequisite.
	 *
	 * @param tag The tag for this prerequisite.
	 * @param parent The owning prerequisite list, if any.
	 */
	public NameLevelPrereq(String tag, PrereqList parent) {
		super(parent);
		mTag = tag;
		mNameCriteria = new StringCriteria(StringCompareType.IS, ""); //$NON-NLS-1$
		mLevelCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 0);
	}

	/**
	 * Loads a prerequisite.
	 *
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 */
	public NameLevelPrereq(PrereqList parent, XMLReader reader) throws IOException {
		this(reader.getName(), parent);
		initializeForLoad();
		String marker = reader.getMarker();
		loadHasAttribute(reader);

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
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
	protected NameLevelPrereq(PrereqList parent, NameLevelPrereq prereq) {
		super(parent, prereq);
		mTag = prereq.mTag;
		mNameCriteria = new StringCriteria(prereq.mNameCriteria);
		mLevelCriteria = new IntegerCriteria(prereq.mLevelCriteria);
	}

	/** Called so that sub-classes can initialize themselves prior to loading. */
	protected void initializeForLoad() {
		// Does nothing
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof NameLevelPrereq && super.equals(obj)) {
			NameLevelPrereq nlp = (NameLevelPrereq) obj;
			return mTag.equals(nlp.mTag) && mNameCriteria.equals(nlp.mNameCriteria) && mLevelCriteria.equals(nlp.mLevelCriteria);
		}
		return false;
	}

	@Override
	public void save(XMLWriter out) {
		out.startTag(mTag);
		saveHasAttribute(out);
		out.finishTagEOL();
		mNameCriteria.save(out, TAG_NAME);
		saveSelf(out);
		if (mLevelCriteria.getType() != NumericCompareType.AT_LEAST || mLevelCriteria.getQualifier() != 0) {
			mLevelCriteria.save(out, TAG_LEVEL);
		}
		out.endTagEOL(mTag, true);
	}

	/** @param reader The XML reader to load from. */
	protected void loadSelf(XMLReader reader) throws IOException {
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
	protected void saveSelf(XMLWriter out) {
		// Does nothing
	}

	/** @return The name comparison object. */
	public StringCriteria getNameCriteria() {
		return mNameCriteria;
	}

	/** @return The level comparison object. */
	public IntegerCriteria getLevelCriteria() {
		return mLevelCriteria;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		ListRow.extractNameables(set, mNameCriteria.getQualifier());
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
	}
}
