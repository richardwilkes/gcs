/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

public class ConditionalModifier extends Bonus {
    public static final String KEY_ROOT      = "conditional_modifier";
    public static final String KEY_SITUATION = "situation";

    private String mSituation;

    public ConditionalModifier() {
        super(1);
        mSituation = "triggering condition";
    }

    public ConditionalModifier(ConditionalModifier other) {
        super(other);
        mSituation = other.mSituation;
    }

    public ConditionalModifier(DataFile dataFile, JsonMap m) throws IOException {
        this();
        loadSelf(dataFile, m);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof ConditionalModifier && super.equals(obj)) {
            return mSituation.equals(((ConditionalModifier) obj).mSituation);
        }
        return false;
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public String getKey() {
        return KEY_ROOT;
    }

    @Override
    public Feature cloneFeature() {
        return new ConditionalModifier(this);
    }

    @Override
    protected void loadSelf(DataFile dataFile, JsonMap m) throws IOException {
        super.loadSelf(dataFile, m);
        setSituation(m.getString(KEY_SITUATION));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_SITUATION, mSituation);
    }

    public String getSituation() {
        return mSituation;
    }

    public void setSituation(String situation) {
        mSituation = situation;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        ListRow.extractNameables(set, mSituation);
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        mSituation = ListRow.nameNameables(map, mSituation);
    }
}
