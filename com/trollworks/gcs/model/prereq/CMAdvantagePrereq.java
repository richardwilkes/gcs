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
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.criteria.CMIntegerCriteria;
import com.trollworks.gcs.model.criteria.CMStringCompareType;
import com.trollworks.gcs.model.criteria.CMStringCriteria;
import com.trollworks.toolkit.io.xml.TKXMLReader;
import com.trollworks.toolkit.io.xml.TKXMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** A (Dis)Advantage prerequisite. */
public class CMAdvantagePrereq extends CMNameLevelPrereq {
	/** The XML tag for this class. */
	public static final String	TAG_ROOT	= "advantage_prereq";	//$NON-NLS-1$
	private static final String	TAG_NOTES	= "notes";				//$NON-NLS-1$
	private static final String	EMPTY		= "";					//$NON-NLS-1$
	private CMStringCriteria	mNotesCriteria;

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public CMAdvantagePrereq(CMPrereqList parent) {
		super(TAG_ROOT, parent);
		mNotesCriteria = new CMStringCriteria(CMStringCompareType.IS_ANYTHING, EMPTY);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public CMAdvantagePrereq(CMPrereqList parent, TKXMLReader reader) throws IOException {
		super(parent, reader);
	}

	private CMAdvantagePrereq(CMPrereqList parent, CMAdvantagePrereq prereq) {
		super(parent, prereq);
		mNotesCriteria = new CMStringCriteria(prereq.mNotesCriteria);
	}

	@Override protected void initializeForLoad() {
		mNotesCriteria = new CMStringCriteria(CMStringCompareType.IS_ANYTHING, EMPTY);
	}

	@Override public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof CMAdvantagePrereq) {
			CMAdvantagePrereq other = (CMAdvantagePrereq) obj;

			return super.equals(other) && mNotesCriteria.equals(other.mNotesCriteria);
		}
		return false;
	}

	@Override protected void loadSelf(TKXMLReader reader) throws IOException {
		if (TAG_NOTES.equals(reader.getName())) {
			mNotesCriteria.load(reader);
		} else {
			super.loadSelf(reader);
		}
	}

	@Override protected void saveSelf(TKXMLWriter out) {
		mNotesCriteria.save(out, TAG_NOTES);
	}

	@Override public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override public CMPrereq clone(CMPrereqList parent) {
		return new CMAdvantagePrereq(parent, this);
	}

	@Override public boolean satisfied(CMCharacter character, CMRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = false;
		CMStringCriteria nameCriteria = getNameCriteria();
		CMIntegerCriteria levelCriteria = getLevelCriteria();

		for (CMAdvantage advantage : character.getAdvantagesIterator()) {
			if (exclude != advantage && nameCriteria.matches(advantage.getName()) && mNotesCriteria.matches(advantage.getNotes())) {
				int levels = advantage.getLevels();

				if (levels < 0) {
					levels = 0;
				}
				satisfied = levelCriteria.matches(levels);
				break;
			}
		}
		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(Msgs.NAME_PART, prefix, has() ? Msgs.HAS : Msgs.DOES_NOT_HAVE, nameCriteria.toString()));
			if (mNotesCriteria.getType() != CMStringCompareType.IS_ANYTHING) {
				builder.append(MessageFormat.format(Msgs.NOTES_PART, mNotesCriteria.toString()));
			}
			builder.append(MessageFormat.format(Msgs.LEVEL_PART, levelCriteria.toString()));
		}
		return satisfied;
	}

	@Override public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		CMRow.extractNameables(set, mNotesCriteria.getQualifier());
	}

	@Override public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mNotesCriteria.setQualifier(CMRow.nameNameables(map, mNotesCriteria.getQualifier()));
	}

	/** @return The notes comparison object. */
	public CMStringCriteria getNotesCriteria() {
		return mNotesCriteria;
	}
}
