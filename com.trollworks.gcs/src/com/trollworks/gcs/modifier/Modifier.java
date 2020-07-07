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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

public abstract class Modifier extends ListRow implements Comparable<Modifier>, HasSourceReference {
    /** The tag for the name. */
    protected static final String  TAG_NAME          = "name";
    /** The tag for the page reference. */
    protected static final String  TAG_REFERENCE     = "reference";
    /** The attribute for whether it is enabled. */
    protected static final String  ATTRIBUTE_ENABLED = "enabled";
    private static final   String  KEY_DISABLED      = "disabled";
    /** The name of the {@link Modifier}. */
    protected              String  mName;
    /** The page reference for the {@link Modifier}. */
    protected              String  mReference;
    protected              boolean mEnabled;
    protected              boolean mReadOnly;

    protected Modifier(DataFile file, Modifier other) {
        super(file, other);
        mName = other.mName;
        mReference = other.mReference;
        mEnabled = other.mEnabled;
    }

    protected Modifier(DataFile file, boolean isContainer) {
        super(file, isContainer);
        mName = getLocalizedName();
        mReference = "";
        mEnabled = !isContainer;
    }

    /** @return An exact clone of this modifier. */
    public abstract Modifier cloneModifier(boolean deep);

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mName = getLocalizedName();
        mReference = "";
        mEnabled = !canHaveChildren();
    }

    @Override
    protected void loadAttributes(XMLReader reader, LoadState state) {
        super.loadAttributes(reader, state);
        mEnabled = !reader.hasAttribute(ATTRIBUTE_ENABLED) || reader.isAttributeSet(ATTRIBUTE_ENABLED);
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (TAG_NAME.equals(name)) {
            mName = reader.readText().replace("\n", " ");
        } else if (TAG_REFERENCE.equals(name)) {
            mReference = reader.readText().replace("\n", " ");
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        mEnabled = !m.getBoolean(KEY_DISABLED);
        mName = m.getString(TAG_NAME);
        mReference = m.getString(TAG_REFERENCE);
    }

    @Override
    protected void saveSelf(JsonWriter w, boolean forUndo) throws IOException {
        w.keyValueNot(KEY_DISABLED, !mEnabled, false);
        w.keyValue(TAG_NAME, mName);
        w.keyValueNot(TAG_REFERENCE, mReference, "");
    }

    public abstract String getNotificationPrefix();

    @Override
    public String getListChangedID() {
        return getNotificationPrefix() + "ListChanged";
    }

    @Override
    public String getRowType() {
        return I18n.Text("Modifier");
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Modifier");
    }

    /** @return The name. */
    public String getName() {
        return mName;
    }

    /**
     * @param name The value to set for name.
     * @return {@code true} if name has changed
     */
    public boolean setName(String name) {
        if (!mName.equals(name)) {
            mName = name;
            notifySingle(getNotificationPrefix() + TAG_NAME);
            return true;
        }
        return false;
    }

    @Override
    public String getReference() {
        return mReference;
    }

    @Override
    public boolean setReference(String reference) {
        if (!mReference.equals(reference)) {
            mReference = reference;
            notifySingle(getNotificationPrefix() + TAG_REFERENCE);
            return true;
        }
        return false;
    }

    @Override
    public String getReferenceHighlight() {
        return getName();
    }

    /** @return The enabled. */
    public boolean isEnabled() {
        return mEnabled;
    }

    /**
     * @param enabled The value to set for enabled.
     * @return {@code true} if enabled has changed.
     */
    public boolean setEnabled(boolean enabled) {
        if (mEnabled != enabled) {
            mEnabled = enabled;
            notifySingle(getNotificationPrefix() + ATTRIBUTE_ENABLED);
            return true;
        }
        return false;
    }

    /** @return Whether this has been marked as "read-only". */
    public boolean isReadOnly() {
        return mReadOnly;
    }

    /** @param readOnly Whether this has been marked as "read-only". */
    public void setReadOnly(boolean readOnly) {
        mReadOnly = readOnly;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Modifier && super.isEquivalentTo(obj)) {
            Modifier row = (Modifier) obj;
            return mEnabled == row.mEnabled && mName.equals(row.mName) && mReference.equals(row.mReference);
        }
        return false;
    }

    @Override
    public String getModifierNotes() {
        return mReadOnly ? I18n.Text("** From container - not modifiable here **") : super.getModifierNotes();
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getName().toLowerCase().contains(text)) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    @Override
    public String toString() {
        return getName();
    }

    /** @return The formatted cost. */
    public abstract String getCostDescription();

    /** @return A full description of this modifier. */
    public abstract String getFullDescription();

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        if (isEnabled()) {
            super.fillWithNameableKeys(set);
            extractNameables(set, mName);
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        if (isEnabled()) {
            super.applyNameableKeys(map);
            mName = nameNameables(map, mName);
        }
    }

    @Override
    public int compareTo(Modifier other) {
        if (this == other) {
            return 0;
        }
        int result = mName.compareTo(other.mName);
        if (result == 0) {
            result = getNotes().compareTo(other.getNotes());
        }
        return result;
    }
}
