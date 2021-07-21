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

import com.trollworks.gcs.utility.I18n;

/** Provides simple storage for the skill level/relative level pair. */
public class SkillLevel {
    /** The skill level. */
    public int    mLevel;
    /** The relative skill level. */
    public int    mRelativeLevel;
    /** The tooltip describing how this level was calculated. */
    public String mToolTip;

    public static String getNoAdditionalModifiers() {
        return I18n.text("No additional modifiers");
    }

    public static String getIncludesModifiersFrom() {
        return I18n.text("Includes modifiers from");
    }

    /**
     * Creates a new SkillLevel.
     *
     * @param level         The skill level.
     * @param relativeLevel The relative skill level.
     */
    public SkillLevel(int level, int relativeLevel) {
        mLevel = level;
        mRelativeLevel = relativeLevel;
        mToolTip = getNoAdditionalModifiers();
    }

    /**
     * Creates a new SkillLevel.
     *
     * @param level         The skill level.
     * @param relativeLevel The relative skill level.
     * @param toolTip       The tooltip to display for this skill.
     */
    public SkillLevel(int level, int relativeLevel, StringBuilder toolTip) {
        this(level, relativeLevel);
        if (toolTip != null && !toolTip.isEmpty()) {
            mToolTip = getIncludesModifiersFrom() + toolTip;
        }
    }

    /** @return The level. */
    public int getLevel() {
        return mLevel;
    }

    /** @return The relativeLevel. */
    public int getRelativeLevel() {
        return mRelativeLevel;
    }

    public boolean isDifferentLevelThan(SkillLevel other) {
        return mLevel != other.mLevel;
    }

    public boolean isDifferentRelativeLevelThan(SkillLevel other) {
        return mRelativeLevel != other.mRelativeLevel;
    }

    public boolean isSameLevelAs(SkillLevel other) {
        return mLevel == other.mLevel && mRelativeLevel == other.mRelativeLevel;
    }

    public String getToolTip() {
        return mToolTip;
    }

    @Override
    public String toString() {
        return "SkillLevel{" + "mLevel=" + mLevel + ", mRelativeLevel=" + mRelativeLevel + '}';
    }
}
