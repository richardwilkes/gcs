/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.Map;
import java.util.Set;

/** An Advantage prerequisite. */
public class AdvantagePrereq extends NameLevelPrereq {
    /** The XML tag for this class. */
    public static final  String         TAG_ROOT  = "advantage_prereq";
    private static final String         TAG_NOTES = "notes";
    private              StringCriteria mNotesCriteria;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public AdvantagePrereq(PrereqList parent) {
        super(TAG_ROOT, parent);
        mNotesCriteria = new StringCriteria(StringCompareType.ANY, "");
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param m      The {@link JsonMap} to load from.
     */
    public AdvantagePrereq(PrereqList parent, JsonMap m) throws IOException {
        super(parent, m);
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
        mNotesCriteria = new StringCriteria(StringCompareType.ANY, "");
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
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mNotesCriteria.load(m.getMap(TAG_NOTES));
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        mNotesCriteria.save(w, TAG_NOTES);
    }

    @Override
    public String getJSONTypeName() {
        return TAG_ROOT;
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
        boolean         satisfied     = false;
        StringCriteria  nameCriteria  = getNameCriteria();
        IntegerCriteria levelCriteria = getLevelCriteria();

        for (Advantage advantage : character.getAdvantagesIterator(false)) {
            if (exclude != advantage && nameCriteria.matches(advantage.getName())) {
                String notes         = advantage.getNotes();
                String modifierNotes = advantage.getModifierNotes();

                if (!modifierNotes.isEmpty()) {
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
            builder.append(MessageFormat.format(I18n.Text("{0}{1} an advantage whose name {2}"), prefix, hasText(), nameCriteria.toString()));
            if (!mNotesCriteria.isTypeAnything()) {
                builder.append(MessageFormat.format(I18n.Text(", notes {0},"), mNotesCriteria.toString()));
            }
            builder.append(MessageFormat.format(I18n.Text(" and level {0}"), levelCriteria.toString()));
        }
        return satisfied;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        ListRow.extractNameables(set, mNotesCriteria.getQualifier());
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        super.applyNameableKeys(map);
        mNotesCriteria.setQualifier(ListRow.nameNameables(map, mNotesCriteria.getQualifier()));
    }

    /** @return The notes comparison object. */
    public StringCriteria getNotesCriteria() {
        return mNotesCriteria;
    }
}
