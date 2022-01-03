/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.notes;

import com.trollworks.gcs.character.CollectedListRow;
import com.trollworks.gcs.character.CollectedOutlines;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;
import javax.swing.Icon;

/** A note. */
public class Note extends CollectedListRow implements HasSourceReference {
    public static final  String KEY_NOTE           = "note";
    public static final  String KEY_NOTE_CONTAINER = "note_container";
    private static final String KEY_TEXT           = "text";
    private static final String KEY_REFERENCE      = "reference";

    private String mText;
    private String mReference;

    /**
     * Creates a new note.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public Note(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
        mText = "";
        mReference = "";
    }

    /**
     * Creates a clone of an existing note and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param note     The note to clone.
     * @param deep     Whether or not to clone the children, grandchildren, etc.
     */
    public Note(DataFile dataFile, Note note, boolean deep) {
        super(dataFile, note);
        mText = note.mText;
        mReference = note.mReference;
        if (deep) {
            int count = note.getChildCount();
            for (int i = 0; i < count; i++) {
                addChild(new Note(dataFile, (Note) note.getChild(i), true));
            }
        }
    }

    public Note(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, m.getString(DataFile.TYPE).equals(KEY_NOTE_CONTAINER));
        load(dataFile, m, state);
    }

    @Override
    public Note cloneRow(DataFile newOwner, boolean deep, boolean forSheet) {
        return new Note(newOwner, this, deep);
    }

    @Override
    public ListOutline getOutlineFromCollectedOutlines(CollectedOutlines outlines) {
        return outlines.getNotesOutline();
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Note row && super.isEquivalentTo(obj)) {
            return mText.equals(row.mText) && mReference.equals(row.mReference);
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.text("Note");
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_NOTE_CONTAINER : KEY_NOTE;
    }

    @Override
    public String getRowType() {
        return I18n.text("Note");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mText = "";
        mReference = "";
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) {
        mText = m.getString(KEY_TEXT);
        mReference = m.getString(KEY_REFERENCE);
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.TYPE);
            if (KEY_NOTE.equals(type) || KEY_NOTE_CONTAINER.equals(type)) {
                addChild(new Note(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        w.keyValueNot(KEY_TEXT, mText, "");
        w.keyValueNot(KEY_REFERENCE, mReference, "");
    }

    /** @return The description. */
    public String getDescription() {
        return mText;
    }

    /**
     * @param description The description to set.
     * @return Whether it was modified.
     */
    public boolean setDescription(String description) {
        if (!mText.equals(description)) {
            mText = description;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getDescription().toLowerCase().contains(text)) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    @Override
    public Object getData(Column column) {
        return NoteColumn.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return NoteColumn.values()[column.getID()].getDataAsText(this);
    }

    @Override
    public String toString() {
        return getDescription();
    }

    @Override
    public Icon getIcon() {
        return FileType.NOTE.getIcon();
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new NoteEditor(this);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        // No nameables
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        // No nameables
    }

    @Override
    public String getReference() {
        return mReference;
    }

    @Override
    public String getReferenceHighlight() {
        return getDescription();
    }

    @Override
    public boolean setReference(String reference) {
        if (!mReference.equals(reference)) {
            mReference = reference;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getToolTip(Column column) {
        return NoteColumn.values()[column.getID()].getToolTip(this);
    }
}
