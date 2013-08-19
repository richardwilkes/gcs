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
import com.trollworks.gcs.model.criteria.CMNumericCriteria;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.toolkit.io.xml.TKXMLNodeType;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** A Spell prerequisite. */
public class CMSpellPrereq extends CMHasPrereq {
	/** The XML tag for this class. */
	public static final String	TAG_ROOT			= "spell_prereq";	//$NON-NLS-1$
	/** The tag/type for name comparision. */
	public static final String	TAG_NAME			= "name";			//$NON-NLS-1$
	/** The tag/type for any. */
	public static final String	TAG_ANY				= "any";			//$NON-NLS-1$
	/** The tag/type for college name comparision. */
	public static final String	TAG_COLLEGE			= "college";		//$NON-NLS-1$
	/** The tag/type for college count comparision. */
	public static final String	TAG_COLLEGE_COUNT	= "college_count";	//$NON-NLS-1$
	private static final String	TAG_QUANTITY		= "quantity";		//$NON-NLS-1$
	private static final String	EMPTY				= "";				//$NON-NLS-1$
	private String				mType;
	private CMStringCriteria	mStringCriteria;
	private CMIntegerCriteria	mQuantityCriteria;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMSpellPrereq(CMPrereqList parent) {
		super(parent);
		mType = TAG_NAME;
		mStringCriteria = new CMStringCriteria(CMStringCompareType.IS, EMPTY);
		mQuantityCriteria = new CMIntegerCriteria(CMNumericCriteria.AT_LEAST, 1);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMSpellPrereq(CMPrereqList parent, TKXMLReader reader) throws IOException {
		this(parent);

		String marker = reader.getMarker();
		loadHasAttribute(reader);

		do {
			if (reader.next() == TKXMLNodeType.START_TAG) {
				String name = reader.getName();

				if (TAG_NAME.equals(name)) {
					setType(TAG_NAME);
					mStringCriteria.load(reader);
				} else if (TAG_ANY.equals(name)) {
					setType(TAG_ANY);
					mQuantityCriteria.load(reader);
				} else if (TAG_COLLEGE.equals(name)) {
					setType(TAG_COLLEGE);
					mStringCriteria.load(reader);
				} else if (TAG_COLLEGE_COUNT.equals(name)) {
					setType(TAG_COLLEGE_COUNT);
					mQuantityCriteria.load(reader);
				} else if (TAG_QUANTITY.equals(name)) {
					mQuantityCriteria.load(reader);
				} else {
					reader.skipTag(name);
				}
			}
		} while (reader.withinMarker(marker));
	}

	/**
	 * Creates a copy of the specified prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param prereq The prerequisite to clone.
	 */
	protected CMSpellPrereq(CMPrereqList parent, CMSpellPrereq prereq) {
		super(parent, prereq);
		mType = prereq.mType;
		mStringCriteria = new CMStringCriteria(prereq.mStringCriteria);
		mQuantityCriteria = new CMIntegerCriteria(prereq.mQuantityCriteria);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMSpellPrereq && super.equals(obj)) {
			CMSpellPrereq other = (CMSpellPrereq) obj;

			return mType.equals(other.mType) && mStringCriteria.equals(other.mStringCriteria) && mQuantityCriteria.equals(other.mQuantityCriteria);
		}
		return false;
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public CMPrereq clone(CMPrereqList parent) {
		return new CMSpellPrereq(parent, this);
	}

	@Override public void save(TKXMLWriter out) {
		out.startTag(TAG_ROOT);
		saveHasAttribute(out);
		out.finishTagEOL();
		if (mType == TAG_NAME || mType == TAG_COLLEGE) {
			mStringCriteria.save(out, mType);
			if (mQuantityCriteria.getType() != CMNumericCriteria.AT_LEAST || mQuantityCriteria.getQualifier() != 1) {
				mQuantityCriteria.save(out, TAG_QUANTITY);
			}
		} else if (mType == TAG_COLLEGE_COUNT) {
			mQuantityCriteria.save(out, mType);
		} else if (mType == TAG_ANY) {
			out.startTag(TAG_ANY);
			out.finishEmptyTagEOL();
			mQuantityCriteria.save(out, TAG_QUANTITY);
		}
		out.endTagEOL(TAG_ROOT, true);
	}

	/** @return The type of comparison to make. */
	public String getType() {
		return mType;
	}

	/**
	 * @param type The type of comparison to make. Must be one of {@link #TAG_NAME},
	 *            {@link #TAG_ANY}, {@link #TAG_COLLEGE}, or {@link #TAG_COLLEGE_COUNT}.
	 */
	public void setType(String type) {
		if (type == TAG_NAME || type == TAG_COLLEGE || type == TAG_COLLEGE_COUNT || type == TAG_ANY) {
			mType = type;
		} else if (TAG_NAME.equals(type)) {
			mType = TAG_NAME;
		} else if (TAG_ANY.equals(type)) {
			mType = TAG_ANY;
		} else if (TAG_COLLEGE.equals(type)) {
			mType = TAG_COLLEGE;
		} else if (TAG_COLLEGE_COUNT.equals(type)) {
			mType = TAG_COLLEGE_COUNT;
		} else {
			mType = TAG_NAME;
		}
	}

	/** @return The string comparison object. */
	public CMStringCriteria getStringCriteria() {
		return mStringCriteria;
	}

	/** @return The quantity comparison object. */
	public CMIntegerCriteria getQuantityCriteria() {
		return mQuantityCriteria;
	}

	@Override public boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix) {
		HashSet<String> colleges = new HashSet<String>();
		String techLevel = null;
		int count = 0;
		boolean satisfied;

		if (exclude instanceof CMSpell) {
			techLevel = ((CMSpell) exclude).getTechLevel();
		}

		for (CMSpell spell : character.getSpellsIterator()) {
			if (exclude != spell && spell.getPoints() > 0) {
				boolean ok;

				if (techLevel != null) {
					String otherTL = spell.getTechLevel();

					ok = otherTL == null || techLevel.equals(otherTL);
				} else {
					ok = true;
				}
				if (ok) {
					if (mType == TAG_NAME) {
						if (mStringCriteria.matches(spell.getName())) {
							count++;
						}
					} else if (mType == TAG_ANY) {
						count++;
					} else if (mType == TAG_COLLEGE) {
						if (mStringCriteria.matches(spell.getCollege())) {
							count++;
						}
					} else if (mType == TAG_COLLEGE_COUNT) {
						colleges.add(spell.getCollege());
					}
				}
			}
		}

		if (mType == TAG_COLLEGE_COUNT) {
			count = colleges.size();
		}

		satisfied = mQuantityCriteria.matches(count);
		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			if (mType == TAG_NAME) {
				builder.append(MessageFormat.format(Msgs.WHOSE_NAME, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? Msgs.ONE_SPELL : Msgs.MULTIPLE_SPELLS, mStringCriteria.toString()));
			} else if (mType == TAG_ANY) {
				builder.append(MessageFormat.format(Msgs.OF_ANY_KIND, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? Msgs.ONE_SPELL : Msgs.MULTIPLE_SPELLS));
			} else if (mType == TAG_COLLEGE) {
				builder.append(MessageFormat.format(Msgs.WHOSE_COLLEGE, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? Msgs.ONE_SPELL : Msgs.MULTIPLE_SPELLS, mStringCriteria.toString()));
			} else if (mType == TAG_COLLEGE_COUNT) {
				builder.append(MessageFormat.format(Msgs.COLLEGE_COUNT, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, mQuantityCriteria.toString()));
			}
		}
		return satisfied;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		if (mType != TAG_COLLEGE_COUNT) {
			CMRow.extractNameables(set, mStringCriteria.getQualifier());
		}
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		if (mType != TAG_COLLEGE_COUNT) {
			mStringCriteria.setQualifier(CMRow.nameNameables(map, mStringCriteria.getQualifier()));
		}
	}
}
