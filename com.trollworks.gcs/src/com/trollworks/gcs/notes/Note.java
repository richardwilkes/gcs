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

package com.trollworks.gcs.notes;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** A note. */
public class Note extends ListRow implements HasSourceReference {
    private static final int    CURRENT_JSON_VERSION = 1;
    /** The XML tag used for items. */
    public static final  String TAG_NOTE             = "note";
    /** The XML tag used for containers. */
    public static final  String TAG_NOTE_CONTAINER   = "note_container";
    private static final String TAG_TEXT             = "text";
    private static final String TAG_REFERENCE        = "reference";
    /** The prefix used in front of all IDs for the notes. */
    public static final  String PREFIX               = GURPSCharacter.CHARACTER_PREFIX + "note.";
    /** The field ID for text changes. */
    public static final  String ID_TEXT              = PREFIX + "Text";
    /** The field ID for page reference changes. */
    public static final  String ID_REFERENCE         = PREFIX + "Reference";
    /** The field ID for when the row hierarchy changes. */
    public static final  String ID_LIST_CHANGED      = PREFIX + "ListChanged";
    private              String mText;
    private              String mReference;

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
        this(dataFile, m.getString(DataFile.KEY_TYPE).equals(TAG_NOTE_CONTAINER));
        load(m, state);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Note && super.isEquivalentTo(obj)) {
            Note row = (Note) obj;
            return mText.equals(row.mText) && mReference.equals(row.mReference);
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Note");
    }

    @Override
    public String getListChangedID() {
        return ID_LIST_CHANGED;
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? TAG_NOTE_CONTAINER : TAG_NOTE;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Note");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mText = "";
        mReference = "";
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) {
        mText = m.getString(TAG_TEXT);
        mReference = m.getString(TAG_REFERENCE);
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (TAG_NOTE.equals(type) || TAG_NOTE_CONTAINER.equals(type)) {
                addChild(new Note(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        w.keyValueNot(TAG_TEXT, mText, "");
        w.keyValueNot(TAG_REFERENCE, mReference, "");
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
            notifySingle(ID_TEXT);
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
    public RetinaIcon getIcon(boolean marker) {
        return marker ? Images.NOT_MARKER : Images.NOT_FILE;
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
            notifySingle(ID_REFERENCE);
            return true;
        }
        return false;
    }

    @Override
    public String getToolTip(Column column) {
        return NoteColumn.values()[column.getID()].getToolTip(this);
    }
}
