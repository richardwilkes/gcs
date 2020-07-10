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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

public class ReactionBonus extends Bonus {
    public static final String TAG_ROOT      = "reaction_bonus";
    public static final String TAG_SITUATION = "situation";
    public static final String REACTION_KEY  = GURPSCharacter.CHARACTER_PREFIX + "reaction";
    private             String mSituation;

    public ReactionBonus() {
        super(1);
        mSituation = "from others";
    }

    public ReactionBonus(ReactionBonus other) {
        super(other);
        mSituation = other.mSituation;
    }

    public ReactionBonus(JsonMap m) throws IOException {
        this();
        loadSelf(m);
    }

    public ReactionBonus(XMLReader reader) throws IOException {
        this();
        load(reader);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof ReactionBonus && super.equals(obj)) {
            return mSituation.equals(((ReactionBonus) obj).mSituation);
        }
        return false;
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
    public String getKey() {
        return REACTION_KEY;
    }

    @Override
    public Feature cloneFeature() {
        return new ReactionBonus(this);
    }

    @Override
    protected void loadSelf(XMLReader reader) throws IOException {
        if (TAG_SITUATION.equals(reader.getName())) {
            setSituation(reader.readText());
        } else {
            super.loadSelf(reader);
        }
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        setSituation(m.getString(TAG_SITUATION));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(TAG_SITUATION, mSituation);
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
