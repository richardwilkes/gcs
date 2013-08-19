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
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A prerequisite list. */
public class CMPrereqList extends CMPrereq {
	/** The XML tag used for the prereq list. */
	public static final String	TAG_ROOT		= "prereq_list";	//$NON-NLS-1$
	private static final String	ATTRIBUTE_ALL	= "all";			//$NON-NLS-1$
	private boolean				mAll;
	private ArrayList<CMPrereq>	mPrereqs;

	/**
	 * Creates a new prerequisite list.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param all Whether only one criteria in this list has to be met, or all of them must be met.
	 */
	public CMPrereqList(CMPrereqList parent, boolean all) {
		super(parent);
		mAll = all;
		mPrereqs = new ArrayList<CMPrereq>();
	}

	/**
	 * Loads a prerequisite list.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMPrereqList(CMPrereqList parent, TKXMLReader reader) throws IOException {
		this(parent, true);

		String marker = reader.getMarker();
		mAll = reader.isAttributeSet(ATTRIBUTE_ALL);

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (TAG_ROOT.equals(name)) {
					mPrereqs.add(new CMPrereqList(this, reader));
				} else if (CMAdvantagePrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new CMAdvantagePrereq(this, reader));
				} else if (CMAttributePrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new CMAttributePrereq(this, reader));
				} else if (CMContainedWeightPrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new CMContainedWeightPrereq(this, reader));
				} else if (CMSkillPrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new CMSkillPrereq(this, reader));
				} else if (CMSpellPrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new CMSpellPrereq(this, reader));
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	/**
	 * Creates a clone of the specified prerequisite list.
	 * 
	 * @param parent The new owning prerequisite list, if any.
	 * @param prereqList The prerequisite to clone.
	 */
	public CMPrereqList(CMPrereqList parent, CMPrereqList prereqList) {
		super(parent);

		mAll = prereqList.mAll;
		mPrereqs = new ArrayList<CMPrereq>(prereqList.mPrereqs.size());
		for (CMPrereq prereq : prereqList.mPrereqs) {
			mPrereqs.add(prereq.clone(this));
		}
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMPrereqList) {
			CMPrereqList other = (CMPrereqList) obj;

			return mAll == other.mAll && mPrereqs.equals(other.mPrereqs);
		}
		return false;
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public void save(TKXMLWriter out) {
		if (!mPrereqs.isEmpty()) {
			out.startTag(TAG_ROOT);
			out.writeAttribute(ATTRIBUTE_ALL, mAll);
			out.finishTagEOL();
			for (CMPrereq prereq : mPrereqs) {
				prereq.save(out);
			}
			out.endTagEOL(TAG_ROOT, true);
		}
	}

	/**
	 * @return Whether only one criteria in this list has to be met, or all of them must be met.
	 */
	public boolean requiresAll() {
		return mAll;
	}

	/**
	 * @param requiresAll Whether only one criteria in this list has to be met, or all of them must
	 *            be met.
	 */
	public void setRequiresAll(boolean requiresAll) {
		mAll = requiresAll;
	}

	/**
	 * @param prereq The prerequisite to work on.
	 * @return The index of the specified prerequisite. -1 will be returned if the component isn't a
	 *         direct child.
	 */
	public int getIndexOf(CMPrereq prereq) {
		return mPrereqs.indexOf(prereq);
	}

	/** @return The number of children in this list. */
	public int getChildCount() {
		return mPrereqs.size();
	}

	/** @return The children of this list. */
	public List<CMPrereq> getChildren() {
		return Collections.unmodifiableList(mPrereqs);
	}

	/**
	 * Adds the specified prerequisite to this list.
	 * 
	 * @param index The index to add the list at.
	 * @param prereq The prerequisite to add.
	 */
	public void add(int index, CMPrereq prereq) {
		mPrereqs.add(index, prereq);
	}

	/**
	 * Removes the specified prerequisite from this list.
	 * 
	 * @param prereq The prerequisite to remove.
	 */
	public void remove(CMPrereq prereq) {
		if (mPrereqs.contains(prereq)) {
			mPrereqs.remove(prereq);
			prereq.mParent = null;
		}
	}

	@Override public boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix) {
		int satisfiedCount = 0;
		int total = mPrereqs.size();
		boolean requiresAll = requiresAll();
		StringBuilder localBuilder = builder != null ? new StringBuilder() : null;
		String localPrefix = "  " + prefix; //$NON-NLS-1$
		boolean satisfied;

		for (CMPrereq prereq : mPrereqs) {
			if (prereq.satisfied(character, exclude, localBuilder, localPrefix)) {
				satisfiedCount++;
			}
		}

		satisfied = satisfiedCount == total || !requiresAll && satisfiedCount > 0;

		if (!satisfied && localBuilder != null && builder != null) {
			builder.append(MessageFormat.format(requiresAll ? Msgs.REQUIRES_ALL : Msgs.REQUIRES_ANY, prefix));
			builder.append(localBuilder.toString());
		}

		return satisfied;
	}

	@Override public CMPrereq clone(CMPrereqList parent) {
		return new CMPrereqList(parent, this);
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		for (CMPrereq prereq : mPrereqs) {
			prereq.fillWithNameableKeys(set);
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		for (CMPrereq prereq : mPrereqs) {
			prereq.applyNameableKeys(map);
		}
	}
}
