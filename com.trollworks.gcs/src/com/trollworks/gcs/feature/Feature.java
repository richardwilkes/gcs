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

import com.trollworks.gcs.utility.xml.XMLWriter;

import java.util.Map;
import java.util.Set;

/** Describes a feature of an advantage, skill, spell, or piece of equipment. */
public interface Feature {
    /** @return The XML tag representing this feature. */
    String getXMLTag();

    /** @return The feature key used in the feature map. */
    String getKey();

    /** @return An exact clone of this feature. */
    Feature cloneFeature();

    /**
     * Saves the feature.
     *
     * @param out The XML writer to use.
     */
    void save(XMLWriter out);

    /** @param set The nameable keys. */
    void fillWithNameableKeys(Set<String> set);

    /** @param map The map of nameable keys to names to apply. */
    void applyNameableKeys(Map<String, String> map);
}
