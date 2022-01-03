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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/** A spell bonus. */
public class SpellBonus extends Bonus {
    public static final  String KEY_ROOT              = "spell_bonus";
    public static final  String KEY_COLLEGE_NAME      = "college_name";
    public static final  String KEY_POWER_SOURCE_NAME = "power_source_name";
    public static final  String KEY_SPELL_NAME        = "spell_name";
    public static final  String KEY_ALL_COLLEGES      = "all_colleges";
    private static final String KEY_CATEGORY          = "category";
    private static final String KEY_MATCH             = "match";
    private static final String KEY_NAME              = "name";

    private boolean        mAllColleges;
    private String         mMatchType;
    private StringCriteria mNameCriteria;
    private StringCriteria mCategoryCriteria;

    /** Creates a new spell bonus. */
    public SpellBonus() {
        super(1);
        mAllColleges = true;
        mMatchType = KEY_COLLEGE_NAME;
        mNameCriteria = new StringCriteria(StringCompareType.IS, "");
        mCategoryCriteria = new StringCriteria(StringCompareType.ANY, "");
    }

    public SpellBonus(DataFile dataFile, JsonMap m) throws IOException {
        this();
        loadSelf(dataFile, m);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public SpellBonus(SpellBonus other) {
        super(other);
        mAllColleges = other.mAllColleges;
        mMatchType = other.mMatchType;
        mNameCriteria = new StringCriteria(other.mNameCriteria);
        mCategoryCriteria = new StringCriteria(other.mCategoryCriteria);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof SpellBonus sb && super.equals(obj)) {
            return mAllColleges == sb.mAllColleges && Objects.equals(mMatchType, sb.mMatchType) && mNameCriteria.equals(sb.mNameCriteria) && mCategoryCriteria.equals(sb.mCategoryCriteria);
        }
        return false;
    }

    @Override
    public Feature cloneFeature() {
        return new SpellBonus(this);
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    public boolean matchesCategories(Set<String> categories) {
        return matchesCategories(mCategoryCriteria, categories);
    }

    @Override
    public String getKey() {
        StringBuilder buffer = new StringBuilder();
        if (mCategoryCriteria.isTypeAnything()) {
            if (mAllColleges) {
                buffer.append(Spell.ID_COLLEGE);
            } else {
                if (KEY_COLLEGE_NAME.equals(mMatchType)) {
                    buffer.append(Spell.ID_COLLEGE);
                } else if (KEY_POWER_SOURCE_NAME.equals(mMatchType)) {
                    buffer.append(Spell.ID_POWER_SOURCE);
                } else {
                    buffer.append(Spell.ID_NAME);
                }
                if (mNameCriteria.isTypeIs()) {
                    buffer.append('/');
                    buffer.append(mNameCriteria.getQualifier());
                } else {
                    buffer.append("*");
                }
            }
        } else {
            buffer.append(Spell.ID_NAME).append("*");
        }
        return buffer.toString();
    }

    @Override
    protected void loadSelf(DataFile dataFile, JsonMap m) throws IOException {
        super.loadSelf(dataFile, m);
        mMatchType = m.getString(KEY_MATCH);
        mAllColleges = KEY_ALL_COLLEGES.equals(mMatchType);
        if (mAllColleges) {
            mMatchType = KEY_COLLEGE_NAME;
        }
        mNameCriteria.load(m.getMap(KEY_NAME));
        mCategoryCriteria.load(m.getMap(KEY_CATEGORY));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_MATCH, mAllColleges ? KEY_ALL_COLLEGES : mMatchType);
        if (!mAllColleges) {
            mNameCriteria.save(w, KEY_NAME);
        }
        mCategoryCriteria.save(w, KEY_CATEGORY);
    }

    /** @return Whether the bonus applies to all colleges. */
    public boolean allColleges() {
        return mAllColleges;
    }

    /** @param all Whether the bonus applies to all colleges. */
    public void allColleges(boolean all) {
        mAllColleges = all;
    }

    /**
     * @return The match type. One of {@link #KEY_COLLEGE_NAME}, {@link #KEY_POWER_SOURCE_NAME}, or
     *         {@link #KEY_SPELL_NAME}.
     */
    public String getMatchType() {
        return mMatchType;
    }

    /** @return The category criteria. */
    public StringCriteria getCategoryCriteria() {
        return mCategoryCriteria;
    }

    public void setMatchType(String type) {
        if (KEY_COLLEGE_NAME.equals(type)) {
            mMatchType = KEY_COLLEGE_NAME;
        } else if (KEY_POWER_SOURCE_NAME.equals(type)) {
            mMatchType = KEY_POWER_SOURCE_NAME;
        } else {
            mMatchType = KEY_SPELL_NAME;
        }
    }

    /** @return The college/spell name criteria. */
    public StringCriteria getNameCriteria() {
        return mNameCriteria;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        if (!mAllColleges) {
            ListRow.extractNameables(set, mNameCriteria.getQualifier());
        }
        ListRow.extractNameables(set, mCategoryCriteria.getQualifier());
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        if (!mAllColleges) {
            mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
        }
        mCategoryCriteria.setQualifier(ListRow.nameNameables(map, mCategoryCriteria.getQualifier()));
    }
}
