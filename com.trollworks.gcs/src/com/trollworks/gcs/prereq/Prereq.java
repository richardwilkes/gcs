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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** The abstract base class prerequisite criteria and prerequisite lists. */
public abstract class Prereq {
    /** The owning prerequisite list, if any. */
    protected PrereqList mParent;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    protected Prereq(PrereqList parent) {
        mParent = parent;
    }

    /** @return The owning prerequisite list, if any. */
    public PrereqList getParent() {
        return mParent;
    }

    /** Removes this prerequisite from its parent. */
    public void removeFromParent() {
        if (mParent != null) {
            mParent.remove(this);
        }
    }

    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The XML tag representing this prereq. */
    public abstract String getXMLTag();

    /**
     * Loads the prerequisite.
     *
     * @param m The {@link JsonMap} to load from.
     * @param state The {@link LoadState} to use.
     */
    public abstract void loadSelf(JsonMap m, LoadState state) throws IOException;

    /**
     * Saves the prerequisite.
     *
     * @param w The {@link JsonWriter} to use.
     */
    public final void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(DataFile.KEY_TYPE, getJSONTypeName());
        saveSelf(w);
        w.endMap();
    }

    /**
     * Saves the prerequisite.
     *
     * @param w The {@link JsonWriter} to use.
     */
    public abstract void saveSelf(JsonWriter w) throws IOException;

    /**
     * @param character The character to check.
     * @param exclude   The data to exclude from the check.
     * @param builder   The {@link StringBuilder} to append this prerequisite's
     *                  satisfied/unsatisfied description to. May be {@code null}.
     * @param prefix    The prefix to add to each line appended to the builder.
     * @return Whether or not this prerequisite is satisfied by the specified character.
     */
    public abstract boolean satisfied(GURPSCharacter character, ListRow exclude, StringBuilder builder, String prefix);

    /**
     * Creates a deep clone of the prerequisite.
     *
     * @param parent The new owning prerequisite list, if any.
     * @return The clone.
     */
    public abstract Prereq clone(PrereqList parent);

    /** @param set The nameable keys. */
    public void fillWithNameableKeys(Set<String> set) {
        // Do nothing by default
    }

    /** @param map The map of nameable keys to names to apply. */
    public void applyNameableKeys(Map<String, String> map) {
        // Do nothing by default
    }
}
