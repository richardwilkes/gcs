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

import com.trollworks.gcs.criteria.StringCompareType;
import com.trollworks.gcs.criteria.StringCriteria;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** A skill bonus. */
public class SkillBonus extends Bonus {
    /** The XML tag. */
    public static final  String             TAG_ROOT           = "skill_bonus";
    private static final String             TAG_SELECTION_TYPE = "selection_type";
    private static final String             TAG_NAME           = "name";
    private static final String             TAG_SPECIALIZATION = "specialization";
    private static final String             TAG_CATEGORY       = "category";
    private              SkillSelectionType mSkillSelectionType;
    private              StringCriteria     mNameCriteria;
    private              StringCriteria     mSpecializationCriteria;
    private              StringCriteria     mCategoryCriteria;

    /** Creates a new skill bonus. */
    public SkillBonus() {
        super(1);
        mSkillSelectionType = SkillSelectionType.SKILLS_WITH_NAME;
        mNameCriteria = new StringCriteria(StringCompareType.IS, "");
        mSpecializationCriteria = new StringCriteria(StringCompareType.ANY, "");
        mCategoryCriteria = new StringCriteria(StringCompareType.ANY, "");
    }

    public SkillBonus(JsonMap m) throws IOException {
        this();
        loadSelf(m);
    }

    /**
     * Loads a {@link SkillBonus}.
     *
     * @param reader The XML reader to use.
     */
    public SkillBonus(XMLReader reader) throws IOException {
        this();
        load(reader);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public SkillBonus(SkillBonus other) {
        super(other);
        mSkillSelectionType = other.mSkillSelectionType;
        mNameCriteria = new StringCriteria(other.mNameCriteria);
        mSpecializationCriteria = new StringCriteria(other.mSpecializationCriteria);
        mCategoryCriteria = new StringCriteria(other.mCategoryCriteria);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof SkillBonus && super.equals(obj)) {
            SkillBonus sb = (SkillBonus) obj;
            if (mNameCriteria.equals(sb.mNameCriteria)) {
                return mSkillSelectionType == sb.mSkillSelectionType && mNameCriteria.equals(sb.mNameCriteria) && mSpecializationCriteria.equals(sb.mSpecializationCriteria) && mCategoryCriteria.equals(sb.mCategoryCriteria);
            }
        }
        return false;
    }

    @Override
    public Feature cloneFeature() {
        return new SkillBonus(this);
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
        switch (mSkillSelectionType) {
        case THIS_WEAPON:
        default:
            return WeaponBonus.THIS_WEAPON_ID;
        case WEAPONS_WITH_NAME:
            return buildKey(WeaponBonus.WEAPON_NAMED_ID_PREFIX);
        case SKILLS_WITH_NAME:
            return buildKey(Skill.ID_NAME);
        }
    }

    private String buildKey(String prefix) {
        StringBuilder buffer = new StringBuilder();
        buffer.append(prefix);
        if (mNameCriteria.isTypeIs() && mSpecializationCriteria.isTypeAnything() && mCategoryCriteria.isTypeAnything()) {
            buffer.append('/');
            buffer.append(mNameCriteria.getQualifier());
        } else {
            buffer.append("*");
        }
        return buffer.toString();
    }

    public boolean matchesCategories(Set<String> categories) {
        return matchesCategories(mCategoryCriteria, categories);
    }

    @Override
    protected void loadSelf(XMLReader reader) throws IOException {
        String name = reader.getName();
        if (TAG_SELECTION_TYPE.equals(name)) {
            mSkillSelectionType = Enums.extract(reader.readText(), SkillSelectionType.values(), SkillSelectionType.SKILLS_WITH_NAME);
        } else if (TAG_NAME.equals(name)) {
            mNameCriteria.load(reader);
        } else if (TAG_SPECIALIZATION.equals(name)) {
            mSpecializationCriteria.load(reader);
        } else if (TAG_CATEGORY.equals(name)) {
            mCategoryCriteria.load(reader);
        } else {
            super.loadSelf(reader);
        }
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        mSkillSelectionType = Enums.extract(m.getString(TAG_SELECTION_TYPE), SkillSelectionType.values(), SkillSelectionType.SKILLS_WITH_NAME);
        if (mSkillSelectionType != SkillSelectionType.THIS_WEAPON) {
            mNameCriteria.load(m.getMap(TAG_NAME));
            mSpecializationCriteria.load(m.getMap(TAG_SPECIALIZATION));
            mCategoryCriteria.load(m.getMap(TAG_CATEGORY));
        }
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(TAG_SELECTION_TYPE, Enums.toId(mSkillSelectionType));
        if (mSkillSelectionType != SkillSelectionType.THIS_WEAPON) {
            mNameCriteria.save(w, TAG_NAME);
            mSpecializationCriteria.save(w, TAG_SPECIALIZATION);
            mCategoryCriteria.save(w, TAG_CATEGORY);
        }
    }

    public SkillSelectionType getSkillSelectionType() {
        return mSkillSelectionType;
    }

    public boolean setSkillSelectionType(SkillSelectionType type) {
        if (mSkillSelectionType != type) {
            mSkillSelectionType = type;
            return true;
        }
        return false;
    }

    /** @return The name criteria. */
    public StringCriteria getNameCriteria() {
        return mNameCriteria;
    }

    /** @return The name criteria. */
    public StringCriteria getSpecializationCriteria() {
        return mSpecializationCriteria;
    }

    /** @return The category criteria. */
    public StringCriteria getCategoryCriteria() {
        return mCategoryCriteria;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        if (mSkillSelectionType != SkillSelectionType.THIS_WEAPON) {
            ListRow.extractNameables(set, mNameCriteria.getQualifier());
            ListRow.extractNameables(set, mSpecializationCriteria.getQualifier());
            ListRow.extractNameables(set, mCategoryCriteria.getQualifier());
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        if (mSkillSelectionType != SkillSelectionType.THIS_WEAPON) {
            mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
            mSpecializationCriteria.setQualifier(ListRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
            mCategoryCriteria.setQualifier(ListRow.nameNameables(map, mCategoryCriteria.getQualifier()));
        }
    }
}
