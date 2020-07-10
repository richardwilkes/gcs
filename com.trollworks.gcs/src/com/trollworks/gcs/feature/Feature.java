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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** Describes a feature of an advantage, skill, spell, or piece of equipment. */
public abstract class Feature {
    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The XML tag representing this feature. */
    public abstract String getXMLTag();

    /** @return The feature key used in the feature map. */
    public abstract String getKey();

    /** @return An exact clone of this feature. */
    public abstract Feature cloneFeature();

    /**
     * Saves the feature.
     *
     * @param w The {@link JsonWriter} to use.
     */
   public final void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(DataFile.KEY_TYPE, getJSONTypeName());
        saveSelf(w);
        w.endMap();
    }

    protected abstract void saveSelf(JsonWriter w) throws IOException;

    /** @param set The nameable keys. */
    public abstract void fillWithNameableKeys(Set<String> set);

    /** @param map The map of nameable keys to names to apply. */
    public abstract void applyNameableKeys(Map<String, String> map);
}
