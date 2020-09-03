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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.CollectedModels;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.nio.file.Path;

/** A template. */
public class Template extends CollectedModels {
    private static final int     CURRENT_JSON_VERSION   = 1;
    private static final String  TAG_ROOT               = "template";
    /** The prefix for all template IDs. */
    public static final  String  TEMPLATE_PREFIX        = "gct.";
    /**
     * The prefix used to indicate a point value is requested from {@link #getValueForID(String)}.
     */
    public static final  String  POINTS_PREFIX          = TEMPLATE_PREFIX + "points.";
    /** The field ID for point total changes. */
    public static final  String  ID_TOTAL_POINTS        = POINTS_PREFIX + "Total";
    /** The field ID for advantage point summary changes. */
    public static final  String  ID_ADVANTAGE_POINTS    = POINTS_PREFIX + "Advantages";
    /** The field ID for disadvantage point summary changes. */
    public static final  String  ID_DISADVANTAGE_POINTS = POINTS_PREFIX + "Disadvantages";
    /** The field ID for quirk point summary changes. */
    public static final  String  ID_QUIRK_POINTS        = POINTS_PREFIX + "Quirks";
    /** The field ID for skill point summary changes. */
    public static final  String  ID_SKILL_POINTS        = POINTS_PREFIX + "Skills";
    /** The field ID for spell point summary changes. */
    public static final  String  ID_SPELL_POINTS        = POINTS_PREFIX + "Spells";
    private              boolean mNeedAdvantagesPointCalculation;
    private              boolean mNeedSkillPointCalculation;
    private              boolean mNeedSpellPointCalculation;
    private              int     mCachedAdvantagePoints;
    private              int     mCachedDisadvantagePoints;
    private              int     mCachedQuirkPoints;
    private              int     mCachedSkillPoints;
    private              int     mCachedSpellPoints;

    /** Creates a new character with only default values set. */
    public Template() {
    }

    /**
     * Creates a new character from the specified file.
     *
     * @param path The path to load the data from.
     * @throws IOException if the data cannot be read or the file doesn't contain a valid character
     *                     sheet.
     */
    public Template(Path path) throws IOException {
        this();
        load(path);
    }

    @Override
    public String getJSONTypeName() {
        return TAG_ROOT;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public FileType getFileType() {
        return FileType.TEMPLATE;
    }

    @Override
    public RetinaIcon getFileIcons() {
        return Images.GCT_FILE;
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        loadModels(m, state);
        calculateAdvantagePoints();
        calculateSkillPoints();
        calculateSpellPoints();
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        saveModels(w, saveType);
    }

    /**
     * @param id The field ID to retrieve the data for.
     * @return The value of the specified field ID, or {@code null} if the field ID is invalid.
     */
    public Object getValueForID(String id) {
        if (ID_ADVANTAGE_POINTS.equals(id)) {
            return Integer.valueOf(getAdvantagePoints());
        } else if (ID_DISADVANTAGE_POINTS.equals(id)) {
            return Integer.valueOf(getDisadvantagePoints());
        } else if (ID_QUIRK_POINTS.equals(id)) {
            return Integer.valueOf(getQuirkPoints());
        } else if (ID_SKILL_POINTS.equals(id)) {
            return Integer.valueOf(getSkillPoints());
        } else if (ID_SPELL_POINTS.equals(id)) {
            return Integer.valueOf(getSpellPoints());
        }
        return null;
    }

    @Override
    protected void startNotifyAtBatchLevelZero() {
        mNeedAdvantagesPointCalculation = false;
        mNeedSkillPointCalculation = false;
        mNeedSpellPointCalculation = false;
    }

    @Override
    public void notify(String type, Object data) {
        super.notify(type, data);
        if (Advantage.ID_POINTS.equals(type) || Advantage.ID_ROUND_COST_DOWN.equals(type) || Advantage.ID_LEVELS.equals(type) || Advantage.ID_LIST_CHANGED.equals(type)) {
            mNeedAdvantagesPointCalculation = true;
        }
        if (Skill.ID_POINTS.equals(type) || Skill.ID_LIST_CHANGED.equals(type)) {
            mNeedSkillPointCalculation = true;
        }
        if (Spell.ID_POINTS.equals(type) || Spell.ID_LIST_CHANGED.equals(type)) {
            mNeedSpellPointCalculation = true;
        }
    }

    @Override
    protected void endNotifyAtBatchLevelOne() {
        if (mNeedAdvantagesPointCalculation) {
            calculateAdvantagePoints();
            notify(ID_ADVANTAGE_POINTS, Integer.valueOf(getAdvantagePoints()));
            notify(ID_DISADVANTAGE_POINTS, Integer.valueOf(getDisadvantagePoints()));
            notify(ID_QUIRK_POINTS, Integer.valueOf(getQuirkPoints()));
        }
        if (mNeedSkillPointCalculation) {
            calculateSkillPoints();
            notify(ID_SKILL_POINTS, Integer.valueOf(getSkillPoints()));
        }
        if (mNeedSpellPointCalculation) {
            calculateSpellPoints();
            notify(ID_SPELL_POINTS, Integer.valueOf(getSpellPoints()));
        }
        if (mNeedAdvantagesPointCalculation || mNeedSkillPointCalculation || mNeedSpellPointCalculation) {
            notify(ID_TOTAL_POINTS, Integer.valueOf(getTotalPoints()));
        }
    }

    private int getTotalPoints() {
        return getAdvantagePoints() + getDisadvantagePoints() + getQuirkPoints() + getSkillPoints() + getSpellPoints();
    }

    /** @return The number of points spent on advantages. */
    public int getAdvantagePoints() {
        return mCachedAdvantagePoints;
    }

    /** @return The number of points spent on disadvantages. */
    public int getDisadvantagePoints() {
        return mCachedDisadvantagePoints;
    }

    /** @return The number of points spent on quirks. */
    public int getQuirkPoints() {
        return mCachedQuirkPoints;
    }

    private void calculateAdvantagePoints() {
        mCachedAdvantagePoints = 0;
        mCachedDisadvantagePoints = 0;
        mCachedQuirkPoints = 0;

        for (Advantage advantage : getAdvantagesIterator(true)) {
            if (!advantage.canHaveChildren()) {
                int pts = advantage.getAdjustedPoints();

                if (pts > 0) {
                    mCachedAdvantagePoints += pts;
                } else if (pts < -1) {
                    mCachedDisadvantagePoints += pts;
                } else if (pts == -1) {
                    mCachedQuirkPoints--;
                }
            }
        }
    }

    /** @return The number of points spent on skills. */
    public int getSkillPoints() {
        return mCachedSkillPoints;
    }

    private void calculateSkillPoints() {
        mCachedSkillPoints = 0;
        for (Skill skill : getSkillsIterator()) {
            if (!skill.canHaveChildren()) {
                mCachedSkillPoints += skill.getPoints();
            }
        }
    }

    /** @return The number of points spent on spells. */
    public int getSpellPoints() {
        return mCachedSpellPoints;
    }

    private void calculateSpellPoints() {
        mCachedSpellPoints = 0;
        for (Spell spell : getSpellsIterator()) {
            if (!spell.canHaveChildren()) {
                mCachedSpellPoints += spell.getPoints();
            }
        }
    }
}
