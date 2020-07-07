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

package com.trollworks.gcs.criteria;

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

public abstract class Criteria {
    protected static final String KEY_QUALIFIER     = "qualifier";
    protected static final String ATTRIBUTE_COMPARE = "compare";

    public abstract void load(JsonMap m) throws IOException;

    public void save(JsonWriter w, String key) throws IOException {
        w.key(key);
        save(w);
    }

    public final void save(JsonWriter w) throws IOException {
        w.startMap();
        saveSelf(w);
        w.endMap();
    }

    protected abstract void saveSelf(JsonWriter w) throws IOException;
}
