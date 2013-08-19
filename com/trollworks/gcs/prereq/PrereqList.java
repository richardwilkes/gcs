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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.widgets.outline.ListRow;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A prerequisite list. */
public class PrereqList extends Prereq {
	private static String		MSG_REQUIRES_ALL;
	private static String		MSG_REQUIRES_ANY;
	/** The XML tag used for the prereq list. */
	public static final String	TAG_ROOT		= "prereq_list";	//$NON-NLS-1$
	private static final String	TAG_WHEN_TL		= "when_tl";		//$NON-NLS-1$
	private static final String	ATTRIBUTE_ALL	= "all";			//$NON-NLS-1$
	private boolean				mAll;
	private IntegerCriteria		mWhenTLCriteria;
	private ArrayList<Prereq>	mPrereqs;

	static {
		LocalizedMessages.initialize(PrereqList.class);
	}

	/**
	 * Creates a new prerequisite list.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param all Whether only one criteria in this list has to be met, or all of them must be met.
	 */
	public PrereqList(PrereqList parent, boolean all) {
		super(parent);
		mAll = all;
		mWhenTLCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, Integer.MIN_VALUE);
		mPrereqs = new ArrayList<Prereq>();
	}

	/**
	 * Loads a prerequisite list.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public PrereqList(PrereqList parent, XMLReader reader) throws IOException {
		this(parent, true);
		String marker = reader.getMarker();
		mAll = reader.isAttributeSet(ATTRIBUTE_ALL);
		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();
				if (TAG_WHEN_TL.equals(name)) {
					mWhenTLCriteria.load(reader);
				} else if (TAG_ROOT.equals(name)) {
					mPrereqs.add(new PrereqList(this, reader));
				} else if (AdvantagePrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new AdvantagePrereq(this, reader));
				} else if (AttributePrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new AttributePrereq(this, reader));
				} else if (ContainedWeightPrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new ContainedWeightPrereq(this, reader));
				} else if (SkillPrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new SkillPrereq(this, reader));
				} else if (SpellPrereq.TAG_ROOT.equals(name)) {
					mPrereqs.add(new SpellPrereq(this, reader));
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
	public PrereqList(PrereqList parent, PrereqList prereqList) {
		super(parent);
		mAll = prereqList.mAll;
		mWhenTLCriteria = new IntegerCriteria(prereqList.mWhenTLCriteria);
		mPrereqs = new ArrayList<Prereq>(prereqList.mPrereqs.size());
		for (Prereq prereq : prereqList.mPrereqs) {
			mPrereqs.add(prereq.clone(this));
		}
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof PrereqList) {
			PrereqList other = (PrereqList) obj;
			return mAll == other.mAll && mWhenTLCriteria.equals(other.mWhenTLCriteria) && mPrereqs.equals(other.mPrereqs);
		}
		return false;
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public void save(XMLWriter out) {
		if (!mPrereqs.isEmpty()) {
			out.startTag(TAG_ROOT);
			out.writeAttribute(ATTRIBUTE_ALL, mAll);
			out.finishTagEOL();
			if (isWhenTLEnabled(mWhenTLCriteria)) {
				mWhenTLCriteria.save(out, TAG_WHEN_TL);
			}
			for (Prereq prereq : mPrereqs) {
				prereq.save(out);
			}
			out.endTagEOL(TAG_ROOT, true);
		}
	}

	/** @return The character's TL criteria. */
	public IntegerCriteria getWhenTLCriteria() {
		return mWhenTLCriteria;
	}

	/**
	 * @param criteria The {@link IntegerCriteria} to check.
	 * @return Whether the character's TL criteria check is enabled.
	 */
	public static boolean isWhenTLEnabled(IntegerCriteria criteria) {
		return criteria.getType() != NumericCompareType.AT_LEAST || criteria.getQualifier() != Integer.MIN_VALUE;
	}

	/**
	 * @param criteria The {@link IntegerCriteria} to work on.
	 * @param enabled Whether the character's TL criteria check is enabled.
	 */
	public static void setWhenTLEnabled(IntegerCriteria criteria, boolean enabled) {
		if (isWhenTLEnabled(criteria) != enabled) {
			criteria.setQualifier(enabled ? 0 : Integer.MIN_VALUE);
		}
	}

	/** @return Whether only one criteria in this list has to be met, or all of them must be met. */
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
	public int getIndexOf(Prereq prereq) {
		return mPrereqs.indexOf(prereq);
	}

	/** @return The number of children in this list. */
	public int getChildCount() {
		return mPrereqs.size();
	}

	/** @return The children of this list. */
	public List<Prereq> getChildren() {
		return Collections.unmodifiableList(mPrereqs);
	}

	/**
	 * Adds the specified prerequisite to this list.
	 * 
	 * @param index The index to add the list at.
	 * @param prereq The prerequisite to add.
	 */
	public void add(int index, Prereq prereq) {
		mPrereqs.add(index, prereq);
	}

	/**
	 * Removes the specified prerequisite from this list.
	 * 
	 * @param prereq The prerequisite to remove.
	 */
	public void remove(Prereq prereq) {
		if (mPrereqs.contains(prereq)) {
			mPrereqs.remove(prereq);
			prereq.mParent = null;
		}
	}

	@Override public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
		if (isWhenTLEnabled(mWhenTLCriteria)) {
			if (!mWhenTLCriteria.matches(NumberUtils.getNonLocalizedInteger(character.getDescription().getTechLevel(), 0))) {
				return true;
			}
		}

		int satisfiedCount = 0;
		int total = mPrereqs.size();
		boolean requiresAll = requiresAll();
		StringBuilder localBuilder = builder != null ? new StringBuilder() : null;
		for (Prereq prereq : mPrereqs) {
			if (prereq.satisfied(character, exclude, localBuilder, prefix)) {
				satisfiedCount++;
			}
		}
		if (localBuilder != null && localBuilder.length() > 0) {
			localBuilder.insert(0, "<ul>"); //$NON-NLS-1$
			localBuilder.append("</ul>"); //$NON-NLS-1$
		}

		boolean satisfied = satisfiedCount == total || !requiresAll && satisfiedCount > 0;
		if (!satisfied && localBuilder != null && builder != null) {
			builder.append(MessageFormat.format(requiresAll ? MSG_REQUIRES_ALL : MSG_REQUIRES_ANY, prefix));
			builder.append(localBuilder.toString());
		}
		return satisfied;
	}

	@Override public Prereq clone(PrereqList parent) {
		return new PrereqList(parent, this);
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		for (Prereq prereq : mPrereqs) {
			prereq.fillWithNameableKeys(set);
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		for (Prereq prereq : mPrereqs) {
			prereq.applyNameableKeys(map);
		}
	}
}
