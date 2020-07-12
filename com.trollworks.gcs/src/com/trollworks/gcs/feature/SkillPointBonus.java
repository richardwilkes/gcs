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

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** A skill point bonus. */
public class SkillPointBonus extends Bonus {
    /** The XML tag. */
    public static final  String         TAG_ROOT           = "skill_point_bonus";
    private static final String         TAG_NAME           = "name";
    private static final String         TAG_SPECIALIZATION = "specialization";
    private static final String         TAG_CATEGORY       = "category";
    private              StringCriteria mNameCriteria;
    private              StringCriteria mSpecializationCriteria;
    private              StringCriteria mCategoryCriteria;

    /** Creates a new skill point bonus. */
    public SkillPointBonus() {
        super(1);
        mNameCriteria = new StringCriteria(StringCompareType.IS, "");
        mSpecializationCriteria = new StringCriteria(StringCompareType.ANY, "");
        mCategoryCriteria = new StringCriteria(StringCompareType.ANY, "");
    }

    public SkillPointBonus(JsonMap m) throws IOException {
        this();
        loadSelf(m);
    }

    /**
     * Creates a clone of the specified skill point bonus.
     *
     * @param other The skill point bonus to clone.
     */
    public SkillPointBonus(SkillPointBonus other) {
        super(other);
        mNameCriteria = new StringCriteria(other.mNameCriteria);
        mSpecializationCriteria = new StringCriteria(other.mSpecializationCriteria);
        mCategoryCriteria = new StringCriteria(other.mCategoryCriteria);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof SkillPointBonus && super.equals(obj)) {
            SkillPointBonus sb = (SkillPointBonus) obj;
            if (mNameCriteria.equals(sb.mNameCriteria)) {
                return mNameCriteria.equals(sb.mNameCriteria) && mSpecializationCriteria.equals(sb.mSpecializationCriteria) && mCategoryCriteria.equals(sb.mCategoryCriteria);
            }
        }
        return false;
    }

    @Override
    public Feature cloneFeature() {
        return new SkillPointBonus(this);
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
        StringBuilder buffer = new StringBuilder();
        buffer.append(Skill.ID_POINTS);
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
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        mNameCriteria.load(m.getMap(TAG_NAME));
        mSpecializationCriteria.load(m.getMap(TAG_SPECIALIZATION));
        mCategoryCriteria.load(m.getMap(TAG_CATEGORY));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        mNameCriteria.save(w, TAG_NAME);
        mSpecializationCriteria.save(w, TAG_SPECIALIZATION);
        mCategoryCriteria.save(w, TAG_CATEGORY);
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
        ListRow.extractNameables(set, mNameCriteria.getQualifier());
        ListRow.extractNameables(set, mSpecializationCriteria.getQualifier());
        ListRow.extractNameables(set, mCategoryCriteria.getQualifier());
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
        mSpecializationCriteria.setQualifier(ListRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
        mCategoryCriteria.setQualifier(ListRow.nameNameables(map, mCategoryCriteria.getQualifier()));
    }

    @Override
    public void addToToolTip(StringBuilder toolTip) {
        if (toolTip != null) {
            toolTip.append("\n").append(getParentName()).append(" [").append(getToolTipAmount()).append(getAmount().getIntegerAmount() == 1 ? " pt]" : " pts]");
        }
    }
}
