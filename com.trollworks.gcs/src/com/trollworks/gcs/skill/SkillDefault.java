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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.Map;
import java.util.Set;

/** Describes a skill default. */
public class SkillDefault {
    public static final String KEY_TYPE           = "type";
    public static final String KEY_NAME           = "name";
    public static final String KEY_SPECIALIZATION = "specialization";
    public static final String KEY_MODIFIER       = "modifier";
    public static final String KEY_LEVEL          = "level";
    public static final String KEY_ADJ_LEVEL      = "adjusted_level";
    public static final String KEY_POINTS         = "points";

    private String mType;
    private String mName;
    private String mSpecialization;
    private int    mModifier;
    private int    mLevel;
    private int    mAdjLevel;
    private int    mPoints;

    /**
     * Creates a new skill default.
     *
     * @param type           The type of default.
     * @param name           The name of the skill to default from. Pass in {@code null} if type is
     *                       not skill-based.
     * @param specialization The specialization of the skill. Pass in {@code null} if this does not
     *                       default from a skill or the skill doesn't require a specialization.
     * @param modifier       The modifier to use.
     */
    public SkillDefault(String type, String name, String specialization, int modifier) {
        setType(type);
        setName(name);
        setSpecialization(specialization);
        setModifier(modifier);
    }

    /**
     * Creates a clone of the specified skill default.
     *
     * @param other The skill default to clone.
     */
    public SkillDefault(SkillDefault other) {
        mType = other.mType;
        mName = other.mName;
        mSpecialization = other.mSpecialization;
        mModifier = other.mModifier;
    }

    /**
     * Creates a skill default.
     *
     * @param m    The {@link JsonMap} to load data from.
     * @param full {@code true} if all fields should be loaded.
     */
    public SkillDefault(JsonMap m, boolean full) {
        setType(m.getString(DataFile.KEY_TYPE));
        mName = m.getString(KEY_NAME);
        mSpecialization = m.getString(KEY_SPECIALIZATION);
        mModifier = m.getInt(KEY_MODIFIER);
        if (full) {
            mLevel = m.getInt(KEY_LEVEL);
            mAdjLevel = m.getInt(KEY_ADJ_LEVEL);
            mPoints = m.getInt(KEY_POINTS);
        }
    }

    /** @return The current level of this default. Temporary storage only. */
    public int getLevel() {
        return mLevel;
    }

    /**
     * @param level Sets the current level of this default. Temporary storage only.
     */
    public void setLevel(int level) {
        mLevel = level;
    }

    /** @return The current level of this default. Temporary storage only. */
    public int getAdjLevel() {
        return mAdjLevel;
    }

    /**
     * @param level Sets the current level of this default. Temporary storage only.
     */
    public void setAdjLevel(int level) {
        mAdjLevel = level;
    }

    /**
     * @return The current points provided by this default. Temporary storage only.
     */
    public int getPoints() {
        return mPoints;
    }

    /**
     * @param points Sets the current points provided by this default. Temporary storage only.
     */
    public void setPoints(int points) {
        mPoints = points;
    }

    /**
     * Saves the skill default.
     *
     * @param w    The {@link JsonWriter} to use.
     * @param full {@code true} if all fields should be saved.
     */
    public void save(JsonWriter w, boolean full) throws IOException {
        w.startMap();
        w.keyValue(DataFile.KEY_TYPE, mType);
        if (SkillDefaultType.isSkillBased(mType)) {
            w.keyValueNot(KEY_NAME, mName, "");
            w.keyValueNot(KEY_SPECIALIZATION, mSpecialization, "");
        }
        w.keyValueNot(KEY_MODIFIER, mModifier, 0);
        if (full) {
            w.keyValue(KEY_LEVEL, mLevel);
            w.keyValue(KEY_ADJ_LEVEL, mAdjLevel);
            w.keyValueNot(KEY_POINTS, mPoints, 0);
        }
        w.endMap();
    }

    /** @return The type of default. */
    public String getType() {
        return mType;
    }

    /** @param type The new type. */
    public void setType(String type) {
        mType = AttributeDef.sanitizeID(type, false);
    }

    /** @return The full name of the skill to default from. */
    public String getFullName(DataFile dataFile) {
        if (SkillDefaultType.isSkillBased(mType)) {
            StringBuilder builder = new StringBuilder();
            builder.append(mName);
            if (!mSpecialization.isEmpty()) {
                builder.append(" (");
                builder.append(mSpecialization);
                builder.append(')');
            }
            if ("parry".equalsIgnoreCase(mType)) {
                builder.append(I18n.Text(" Parry"));
            } else if ("block".equalsIgnoreCase(mType)) {
                builder.append(I18n.Text(" Block"));
            }
            return builder.toString();
        }
        return Skill.resolveAttributeName(dataFile, mType);
    }

    /**
     * @return The name of the skill to default from. Only valid when {@link #getType()} returns a
     *         {@link SkillDefaultType} which is skill based.
     */
    public String getName() {
        return mName;
    }

    /** @param name The new name. */
    public void setName(String name) {
        mName = name != null ? name : "";
    }

    /**
     * @return The specialization of the skill to default from. Only valid when {@link #getType()}
     *         returns a {@link SkillDefaultType} which is skill based.
     */
    public String getSpecialization() {
        return mSpecialization;
    }

    /** @param specialization The new specialization. */
    public void setSpecialization(String specialization) {
        mSpecialization = specialization != null ? specialization : "";
    }

    /** @return The modifier. */
    public int getModifier() {
        return mModifier;
    }

    /** @param modifier The new modifier. */
    public void setModifier(int modifier) {
        mModifier = modifier;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof SkillDefault) {
            SkillDefault sd = (SkillDefault) obj;
            return mModifier == sd.mModifier && mType.equals(sd.mType) && mName.equals(sd.mName) && mSpecialization.equals(sd.mSpecialization);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    /** @param set The nameable keys. */
    public void fillWithNameableKeys(Set<String> set) {
        ListRow.extractNameables(set, getName());
        ListRow.extractNameables(set, getSpecialization());
    }

    /** @param map The map of nameable keys to names to apply. */
    public void applyNameableKeys(Map<String, String> map) {
        setName(ListRow.nameNameables(map, getName()));
        setSpecialization(ListRow.nameNameables(map, getSpecialization()));
    }

    public String getModifierAsString() {
        if (mModifier > 0) {
            return " + " + mModifier;
        } else if (mModifier < 0) {
            return " - " + -mModifier;
        }
        return "";
    }
}
