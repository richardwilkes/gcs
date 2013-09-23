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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import static com.trollworks.gcs.prereq.HasPrereq_LS.*;
import static com.trollworks.gcs.prereq.SpellPrereq_LS.*;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.xml.XMLNodeType;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

@Localized({
				@LS(key = "ONE_SPELL", msg = "spell"),
				@LS(key = "MULTIPLE_SPELLS", msg = "spells"),
				@LS(key = "WHOSE_NAME", msg = "{0}{1} {2} {3} whose name {4}\n"),
				@LS(key = "OF_ANY_KIND", msg = "{0}{1} {2} {3} of any kind\n"),
				@LS(key = "WHOSE_COLLEGE", msg = "{0}{1} {2} {3} whose college {4}\n"),
				@LS(key = "COLLEGE_COUNT", msg = "{0}{1} college count which {2}\n"),
})
/** A Spell prerequisite. */
public class SpellPrereq extends HasPrereq {
	/** The XML tag for this class. */
	public static final String	TAG_ROOT			= "spell_prereq";	//$NON-NLS-1$
	/** The tag/type for name comparison. */
	public static final String	TAG_NAME			= "name";			//$NON-NLS-1$
	/** The tag/type for any. */
	public static final String	TAG_ANY				= "any";			//$NON-NLS-1$
	/** The tag/type for college name comparison. */
	public static final String	TAG_COLLEGE			= "college";		//$NON-NLS-1$
	/** The tag/type for college count comparison. */
	public static final String	TAG_COLLEGE_COUNT	= "college_count";	//$NON-NLS-1$
	private static final String	TAG_QUANTITY		= "quantity";		//$NON-NLS-1$
	private static final String	EMPTY				= "";				//$NON-NLS-1$
	private String				mType;
	private StringCriteria		mStringCriteria;
	private IntegerCriteria		mQuantityCriteria;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public SpellPrereq(PrereqList parent) {
		super(parent);
		mType = TAG_NAME;
		mStringCriteria = new StringCriteria(StringCompareType.IS, EMPTY);
		mQuantityCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 1);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 */
	public SpellPrereq(PrereqList parent, XMLReader reader) throws IOException {
		this(parent);

		String marker = reader.getMarker();
		loadHasAttribute(reader);

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
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
	protected SpellPrereq(PrereqList parent, SpellPrereq prereq) {
		super(parent, prereq);
		mType = prereq.mType;
		mStringCriteria = new StringCriteria(prereq.mStringCriteria);
		mQuantityCriteria = new IntegerCriteria(prereq.mQuantityCriteria);
	}

	@Override
	public boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof SpellPrereq && super.equals(obj)) {
			SpellPrereq sp = (SpellPrereq) obj;
			return mType.equals(sp.mType) && mStringCriteria.equals(sp.mStringCriteria) && mQuantityCriteria.equals(sp.mQuantityCriteria);
		}
		return false;
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public Prereq clone(PrereqList parent) {
		return new SpellPrereq(parent, this);
	}

	@Override
	public void save(XMLWriter out) {
		out.startTag(TAG_ROOT);
		saveHasAttribute(out);
		out.finishTagEOL();
		if (mType == TAG_NAME || mType == TAG_COLLEGE) {
			mStringCriteria.save(out, mType);
			if (mQuantityCriteria.getType() != NumericCompareType.AT_LEAST || mQuantityCriteria.getQualifier() != 1) {
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
	public StringCriteria getStringCriteria() {
		return mStringCriteria;
	}

	/** @return The quantity comparison object. */
	public IntegerCriteria getQuantityCriteria() {
		return mQuantityCriteria;
	}

	@Override
	public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
		HashSet<String> colleges = new HashSet<>();
		String techLevel = null;
		int count = 0;
		boolean satisfied;

		if (exclude instanceof Spell) {
			techLevel = ((Spell) exclude).getTechLevel();
		}

		for (Spell spell : character.getSpellsIterator()) {
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
				builder.append(MessageFormat.format(WHOSE_NAME, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? ONE_SPELL : MULTIPLE_SPELLS, mStringCriteria.toString()));
			} else if (mType == TAG_ANY) {
				builder.append(MessageFormat.format(OF_ANY_KIND, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? ONE_SPELL : MULTIPLE_SPELLS));
			} else if (mType == TAG_COLLEGE) {
				builder.append(MessageFormat.format(WHOSE_COLLEGE, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString(EMPTY), mQuantityCriteria.getQualifier() == 1 ? ONE_SPELL : MULTIPLE_SPELLS, mStringCriteria.toString()));
			} else if (mType == TAG_COLLEGE_COUNT) {
				builder.append(MessageFormat.format(COLLEGE_COUNT, prefix, has() ? HAS : DOES_NOT_HAVE, mQuantityCriteria.toString()));
			}
		}
		return satisfied;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		if (mType != TAG_COLLEGE_COUNT) {
			ListRow.extractNameables(set, mStringCriteria.getQualifier());
		}
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		if (mType != TAG_COLLEGE_COUNT) {
			mStringCriteria.setQualifier(ListRow.nameNameables(map, mStringCriteria.getQualifier()));
		}
	}
}
