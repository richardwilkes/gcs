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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashMap;
import java.util.HashSet;

/** An Advantage prerequisite. */
public class AdvantagePrereq extends NameLevelPrereq {
	@Localize("{0}{1} an advantage whose name {2}")
	@Localize(locale = "de", value = "{0}{1} einen Vorteil, dessen Name {2}")
	@Localize(locale = "ru", value = "{0}{1}преимущество с названием {2}")
	private static String		NAME_PART;
	@Localize(", notes {0},")
	@Localize(locale = "de", value = ", Notizen {0},")
	@Localize(locale = "ru", value = ", заметок {0},")
	private static String		NOTES_PART;
	@Localize(" and level {0}")
	@Localize(locale = "de", value = " und Stufe {0}")
	@Localize(locale = "ru", value = " и уровень {0}\n ")
	private static String		LEVEL_PART;

	static {
		Localization.initialize();
	}

	/** The XML tag for this class. */
	public static final String	TAG_ROOT	= "advantage_prereq";	//$NON-NLS-1$
	private static final String	TAG_NOTES	= "notes";				//$NON-NLS-1$
	private static final String	EMPTY		= "";					//$NON-NLS-1$
	private StringCriteria		mNotesCriteria;

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
		if (obj == this) {
			return true;
		}
		if (obj instanceof AdvantagePrereq && super.equals(obj)) {
			return mNotesCriteria.equals(((AdvantagePrereq) obj).mNotesCriteria);
		}
		return false;
	}

	@Override
	public int hashCode() {
		return super.hashCode();
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
			builder.append(MessageFormat.format(NAME_PART, prefix, has() ? HasPrereq.HAS : HasPrereq.DOES_NOT_HAVE, nameCriteria.toString()));
			if (mNotesCriteria.getType() != StringCompareType.IS_ANYTHING) {
				builder.append(MessageFormat.format(NOTES_PART, mNotesCriteria.toString()));
			}
			builder.append(MessageFormat.format(LEVEL_PART, levelCriteria.toString()));
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
