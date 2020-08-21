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

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/**
 * An abstract prerequisite class for comparison of name and level and whether or not the specific
 * item is present.
 */
public abstract class NameLevelPrereq extends HasPrereq {
    /** Provided for sub-classes. */
    private static final String          TAG_NAME  = "name";
    private static final String          TAG_LEVEL = "level";
    private              String          mTag;
    private              StringCriteria  mNameCriteria;
    private              IntegerCriteria mLevelCriteria;

    /**
     * Creates a new prerequisite.
     *
     * @param tag    The tag for this prerequisite.
     * @param parent The owning prerequisite list, if any.
     */
    public NameLevelPrereq(String tag, PrereqList parent) {
        super(parent);
        mTag = tag;
        mNameCriteria = new StringCriteria(StringCompareType.IS, "");
        mLevelCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 0);
    }

    /**
     * Loads a prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param m      The {@link JsonMap} to load from.
     */
    public NameLevelPrereq(PrereqList parent, JsonMap m) throws IOException {
        this(m.getString(DataFile.KEY_TYPE), parent);
        initializeForLoad();
        loadSelf(m, new LoadState());
    }

    /**
     * Creates a copy of the specified prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param prereq The prerequisite to clone.
     */
    protected NameLevelPrereq(PrereqList parent, NameLevelPrereq prereq) {
        super(parent, prereq);
        mTag = prereq.mTag;
        mNameCriteria = new StringCriteria(prereq.mNameCriteria);
        mLevelCriteria = new IntegerCriteria(prereq.mLevelCriteria);
    }

    /** Called so that sub-classes can initialize themselves prior to loading. */
    protected void initializeForLoad() {
        // Does nothing
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof NameLevelPrereq && super.equals(obj)) {
            NameLevelPrereq nlp = (NameLevelPrereq) obj;
            return mTag.equals(nlp.mTag) && mNameCriteria.equals(nlp.mNameCriteria) && mLevelCriteria.equals(nlp.mLevelCriteria);
        }
        return false;
    }

    @Override
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mNameCriteria.load(m.getMap(TAG_NAME));
        if (m.has(TAG_LEVEL)) {
            mLevelCriteria.load(m.getMap(TAG_LEVEL));
        }
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        mNameCriteria.save(w, TAG_NAME);
        if (mLevelCriteria.getType() != NumericCompareType.AT_LEAST || mLevelCriteria.getQualifier() != 0) {
            mLevelCriteria.save(w, TAG_LEVEL);
        }
    }

    /** @return The name comparison object. */
    public StringCriteria getNameCriteria() {
        return mNameCriteria;
    }

    /** @return The level comparison object. */
    public IntegerCriteria getLevelCriteria() {
        return mLevelCriteria;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        ListRow.extractNameables(set, mNameCriteria.getQualifier());
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
    }
}
