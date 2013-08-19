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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.xml.XMLReader;
import com.trollworks.ttk.xml.XMLWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** An Advantage prerequisite. */
public class AdvantagePrereq extends NameLevelPrereq {
	private static String		MSG_NAME_PART;
	private static String		MSG_NOTES_PART;
	/** The XML tag for this class. */
	public static final String	TAG_ROOT	= "advantage_prereq";	//$NON-NLS-1$
	private static final String	TAG_NOTES	= "notes";				//$NON-NLS-1$
	private static final String	EMPTY		= "";					//$NON-NLS-1$
	private StringCriteria		mNotesCriteria;

	static {
		LocalizedMessages.initialize(AdvantagePrereq.class);
	}

	/**
	 * Creates a new prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 */
	public AdvantagePrereq(PrereqList parent) {
		super(TAG_ROOT, parent);
		mNotesCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, EMPTY);
	}

	/**
	 * Loads a prerequisite.
	 * 
	 * @param parent The owning prerequisite list, if any.
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public AdvantagePrereq(PrereqList parent, XMLReader reader) throws IOException {
		super(parent, reader);
	}

	private AdvantagePrereq(PrereqList parent, AdvantagePrereq prereq) {
		super(parent, prereq);
		mNotesCriteria = new StringCriteria(prereq.mNotesCriteria);
	}

	@Override
	protected void initializeForLoad() {
		mNotesCriteria = new StringCriteria(StringCompareType.IS_ANYTHING, EMPTY);
	}

	@Override
	public boolean equals(Object obj) {
		if (this == obj) {
			return true;
		}
		if (obj instanceof AdvantagePrereq) {
			AdvantagePrereq other = (AdvantagePrereq) obj;

			return super.equals(other) && mNotesCriteria.equals(other.mNotesCriteria);
		}
		return false;
	}

	@Override
	protected void loadSelf(XMLReader reader) throws IOException {
		if (TAG_NOTES.equals(reader.getName())) {
			mNotesCriteria.load(reader);
		} else {
			super.loadSelf(reader);
		}
	}

	@Override
	protected void saveSelf(XMLWriter out) {
		mNotesCriteria.save(out, TAG_NOTES);
	}

	@Override
	public String getXMLTag() {
		return TAG_ROOT;
	}

	@Override
	public Prereq clone(PrereqList parent) {
		return new AdvantagePrereq(parent, this);
	}

	@Override
	public boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix) {
		boolean satisfied = false;
		StringCriteria nameCriteria = getNameCriteria();
		IntegerCriteria levelCriteria = getLevelCriteria();

		for (Advantage advantage : character.getAdvantagesIterator()) {
			if (exclude != advantage && nameCriteria.matches(advantage.getName())) {
				String notes = advantage.getNotes();
				String modifierNotes = advantage.getModifierNotes();

				if (modifierNotes.length() > 0) {
					notes = modifierNotes + '\n' + notes;
				}
				if (mNotesCriteria.matches(notes)) {
					int levels = advantage.getLevels();
					if (levels < 0) {
						levels = 0;
					}
					satisfied = levelCriteria.matches(levels);
					break;
				}
			}
		}
		if (!has()) {
			satisfied = !satisfied;
		}
		if (!satisfied && builder != null) {
			builder.append(MessageFormat.format(MSG_NAME_PART, prefix, has() ? MSG_HAS : MSG_DOES_NOT_HAVE, nameCriteria.toString()));
			if (mNotesCriteria.getType() != StringCompareType.IS_ANYTHING) {
				builder.append(MessageFormat.format(MSG_NOTES_PART, mNotesCriteria.toString()));
			}
			builder.append(MessageFormat.format(MSG_LEVEL_PART, levelCriteria.toString()));
		}
		return satisfied;
	}

	@Override
	public void fillWithNameableKeys(HashSet<String> set) {
		super.fillWithNameableKeys(set);
		ListRow.extractNameables(set, mNotesCriteria.getQualifier());
	}

	@Override
	public void applyNameableKeys(HashMap<String, String> map) {
		super.applyNameableKeys(map);
		mNotesCriteria.setQualifier(ListRow.nameNameables(map, mNotesCriteria.getQualifier()));
	}

	/** @return The notes comparison object. */
	public StringCriteria getNotesCriteria() {
		return mNotesCriteria;
	}
}
