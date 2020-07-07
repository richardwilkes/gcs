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

import com.trollworks.gcs.criteria.IntegerCriteria;
import com.trollworks.gcs.criteria.NumericCompareType;
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

/** A weapon bonus. */
public class WeaponBonus extends Bonus {
    public static final  String              THIS_WEAPON_ID         = "\u0001";
    public static final  String              WEAPON_NAMED_ID_PREFIX = "weapon_named.";
    /** The XML tag. */
    public static final  String              TAG_ROOT               = "weapon_bonus";
    private static final String              TAG_SELECTION_TYPE     = "selection_type";
    private static final String              TAG_NAME               = "name";
    private static final String              TAG_SPECIALIZATION     = "specialization";
    private static final String              TAG_LEVEL              = "level";
    private static final String              TAG_CATEGORY           = "category";
    private static final String              TAG_PARENT_ONLY        = "parent_only";
    private              WeaponSelectionType mWeaponSelectionType;
    private              StringCriteria      mNameCriteria;
    private              StringCriteria      mSpecializationCriteria;
    private              IntegerCriteria     mLevelCriteria;
    private              StringCriteria      mCategoryCriteria;

    /** Creates a new skill bonus. */
    public WeaponBonus() {
        super(1);
        mWeaponSelectionType = WeaponSelectionType.WEAPONS_WITH_REQUIRED_SKILL;
        mNameCriteria = new StringCriteria(StringCompareType.IS, "");
        mSpecializationCriteria = new StringCriteria(StringCompareType.ANY, "");
        mLevelCriteria = new IntegerCriteria(NumericCompareType.AT_LEAST, 0);
        mCategoryCriteria = new StringCriteria(StringCompareType.ANY, "");
    }

    public WeaponBonus(JsonMap m) throws IOException {
        this();
        loadSelf(m);
    }

    /**
     * Loads a {@link WeaponBonus}.
     *
     * @param reader The XML reader to use.
     */
    public WeaponBonus(XMLReader reader) throws IOException {
        this();
        load(reader);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public WeaponBonus(WeaponBonus other) {
        super(other);
        mWeaponSelectionType = other.mWeaponSelectionType;
        mNameCriteria = new StringCriteria(other.mNameCriteria);
        mSpecializationCriteria = new StringCriteria(other.mSpecializationCriteria);
        mLevelCriteria = new IntegerCriteria(other.mLevelCriteria);
        mCategoryCriteria = new StringCriteria(other.mCategoryCriteria);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof WeaponBonus && super.equals(obj)) {
            WeaponBonus wb = (WeaponBonus) obj;
            return mWeaponSelectionType == wb.mWeaponSelectionType && mNameCriteria.equals(wb.mNameCriteria) && mSpecializationCriteria.equals(wb.mSpecializationCriteria) && mLevelCriteria.equals(wb.mLevelCriteria) && mCategoryCriteria.equals(wb.mCategoryCriteria);
        }
        return false;
    }

    @Override
    public Feature cloneFeature() {
        return new WeaponBonus(this);
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
        switch (mWeaponSelectionType) {
        case THIS_WEAPON:
        default:
            return THIS_WEAPON_ID;
        case WEAPONS_WITH_NAME:
            return buildKey(WEAPON_NAMED_ID_PREFIX);
        case WEAPONS_WITH_REQUIRED_SKILL:
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
            mWeaponSelectionType = Enums.extract(reader.readText(), WeaponSelectionType.values(), WeaponSelectionType.WEAPONS_WITH_REQUIRED_SKILL);
        } else if (TAG_NAME.equals(name)) {
            mNameCriteria.load(reader);
        } else if (TAG_SPECIALIZATION.equals(name)) {
            mSpecializationCriteria.load(reader);
        } else if (TAG_LEVEL.equals(name)) {
            mLevelCriteria.load(reader);
        } else if (TAG_CATEGORY.equals(name)) {
            mCategoryCriteria.load(reader);
        } else {
            super.loadSelf(reader);
        }
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        mWeaponSelectionType = Enums.extract(m.getString(TAG_SELECTION_TYPE), WeaponSelectionType.values(), WeaponSelectionType.WEAPONS_WITH_REQUIRED_SKILL);
        switch (mWeaponSelectionType) {
        case THIS_WEAPON:
        default:
            break;
        case WEAPONS_WITH_NAME:
            mNameCriteria.load(m.getMap(TAG_NAME));
            mSpecializationCriteria.load(m.getMap(TAG_SPECIALIZATION));
            mCategoryCriteria.load(m.getMap(TAG_CATEGORY));
            break;
        case WEAPONS_WITH_REQUIRED_SKILL:
            mNameCriteria.load(m.getMap(TAG_NAME));
            mSpecializationCriteria.load(m.getMap(TAG_SPECIALIZATION));
            mLevelCriteria.load(m.getMap(TAG_LEVEL));
            mCategoryCriteria.load(m.getMap(TAG_CATEGORY));
            break;
        }
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(TAG_SELECTION_TYPE, Enums.toId(mWeaponSelectionType));
        switch (mWeaponSelectionType) {
        case THIS_WEAPON:
        default:
            break;
        case WEAPONS_WITH_NAME:
            mNameCriteria.save(w, TAG_NAME);
            mSpecializationCriteria.save(w, TAG_SPECIALIZATION);
            mCategoryCriteria.save(w, TAG_CATEGORY);
            break;
        case WEAPONS_WITH_REQUIRED_SKILL:
            mNameCriteria.save(w, TAG_NAME);
            mSpecializationCriteria.save(w, TAG_SPECIALIZATION);
            mLevelCriteria.save(w, TAG_LEVEL);
            mCategoryCriteria.save(w, TAG_CATEGORY);
            break;
        }
    }

    public WeaponSelectionType getWeaponSelectionType() {
        return mWeaponSelectionType;
    }

    public boolean setWeaponSelectionType(WeaponSelectionType type) {
        if (mWeaponSelectionType != type) {
            mWeaponSelectionType = type;
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

    /** @return The level criteria. */
    public IntegerCriteria getLevelCriteria() {
        return mLevelCriteria;
    }

    /** @return The category criteria. */
    public StringCriteria getCategoryCriteria() {
        return mCategoryCriteria;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        switch (mWeaponSelectionType) {
        case THIS_WEAPON:
        default:
            break;
        case WEAPONS_WITH_NAME:
        case WEAPONS_WITH_REQUIRED_SKILL:
            ListRow.extractNameables(set, mNameCriteria.getQualifier());
            ListRow.extractNameables(set, mSpecializationCriteria.getQualifier());
            ListRow.extractNameables(set, mCategoryCriteria.getQualifier());
            break;
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        switch (mWeaponSelectionType) {
        case THIS_WEAPON:
        default:
            break;
        case WEAPONS_WITH_NAME:
        case WEAPONS_WITH_REQUIRED_SKILL:
            mNameCriteria.setQualifier(ListRow.nameNameables(map, mNameCriteria.getQualifier()));
            mSpecializationCriteria.setQualifier(ListRow.nameNameables(map, mSpecializationCriteria.getQualifier()));
            mCategoryCriteria.setQualifier(ListRow.nameNameables(map, mCategoryCriteria.getQualifier()));
            break;
        }
    }

    @Override
    public String getToolTipAmount() {
        return getAmount().getAmountAsWeaponBonus();
    }
}
