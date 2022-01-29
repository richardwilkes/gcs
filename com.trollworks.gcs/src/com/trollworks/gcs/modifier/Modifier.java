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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

public abstract class Modifier extends ListRow implements Comparable<Modifier>, HasSourceReference {
    protected static final String KEY_NAME      = "name";
    protected static final String KEY_REFERENCE = "reference";
    private static final   String KEY_DISABLED  = "disabled";

    protected String  mName;
    protected String  mReference;
    protected boolean mEnabled;
    protected boolean mReadOnly;

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
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        mEnabled = !m.getBoolean(KEY_DISABLED);
        mName = m.getString(KEY_NAME);
        mReference = m.getString(KEY_REFERENCE);
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        w.keyValueNot(KEY_DISABLED, !mEnabled, false);
        w.keyValue(KEY_NAME, mName);
        w.keyValueNot(KEY_REFERENCE, mReference, "");
    }

    @Override
    public String getRowType() {
        return I18n.text("Modifier");
    }

    @Override
    public String getLocalizedName() {
        return I18n.text("Modifier");
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
            notifyOfChange();
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
            notifyOfChange();
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
            notifyOfChange();
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
        if (obj instanceof Modifier row && super.isEquivalentTo(obj)) {
            return mEnabled == row.mEnabled && mName.equals(row.mName) && mReference.equals(row.mReference);
        }
        return false;
    }

    @Override
    public String getModifierNotes() {
        return mReadOnly ? I18n.text("** From container - not modifiable here **") : super.getModifierNotes();
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
