/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** Provides simple storage for the skill level/relative level pair. */
public class SkillLevel {
    @Localize("Includes modifiers from")
    @Localize(locale = "de", value = "Enthält Modifikatoren von")
    @Localize(locale = "ru", value = "Включает в себя модификаторы из")
    @Localize(locale = "es", value = "Incluye modificadores de")
    static String INCLUDES;
    @Localize("No additional modifiers")
    @Localize(locale = "de", value = "Keine zusätzlichen Modifikatoren")
    @Localize(locale = "ru", value = "Никаких дополнительных модификаторов")
    @Localize(locale = "es", value = "No hay modificadores adicionales")
    static String NO_MODIFIERS;

    static {
        Localization.initialize();
    }

    /** The skill level. */
    public int    mLevel;
    /** The relative skill level. */
    public int    mRelativeLevel;
    /** The tooltip describing how this level was calculated. */
    public String mToolTip;

    public String getToolTip() {
        return mToolTip;
    }

    /**
     * Creates a new {@link SkillLevel}.
     *
     * @param level         The skill level.
     * @param relativeLevel The relative skill level.
     */
    public SkillLevel(int level, int relativeLevel) {
        mLevel         = level;
        mRelativeLevel = relativeLevel;
        mToolTip       = NO_MODIFIERS;
    }

    /**
     * Creates a new {@link SkillLevel}.
     *
     * @param level         The skill level.
     * @param relativeLevel The relative skill level.
     * @param toolTip       The tooltip to display for this skill.
     */
    public SkillLevel(int level, int relativeLevel, StringBuilder toolTip) {
        mLevel         = level;
        mRelativeLevel = relativeLevel;
        mToolTip       = NO_MODIFIERS;
        if (toolTip != null && toolTip.length() > 0) {
            mToolTip = INCLUDES + toolTip.toString();
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

}
