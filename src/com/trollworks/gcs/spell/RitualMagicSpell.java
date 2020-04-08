/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.utility.I18n;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashSet;
import java.util.Set;

/**
 * A GURPS Spell for the Ritual Magic system.
 * <p>
 * Ritual Magic spells are techniques and as such they must default to a skill. The convention used
 * is to default to a skill named by RitualMagicSpell.getDefaultSkillName() whose specialization is
 * the actual college of the spell.
 */
public class RitualMagicSpell extends Spell {
    private static final int    CURRENT_VERSION        = 1;
    /** The XML tag used for items. */
    public static final  String TAG_RITUAL_MAGIC_SPELL = "ritual_magic_spell";
    /**
     * The (positive) number of spell prerequsites needed to cast this spell. Used as skill penalty
     * relative to the College skill
     */
    private              int    mPrerequisiteSpellsCount;

    /**
     * Creates a new ritual magic spell.
     *
     * @param dataFile The data file to associate it with.
     */
    public RitualMagicSpell(DataFile dataFile) {
        super(dataFile, false);
        mPoints = 0;
        updateLevel(false);
    }

    /**
     * Creates a clone of an existing ritual magic spell and associates it with the specified data
     * file.
     *
     * @param dataFile         The data file to associate it with.
     * @param ritualMagicSpell The spell to clone.
     * @param deep             Whether or not to clone the children, grandchildren, etc.
     * @param forSheet         Whether this is for a character sheet or a list.
     */
    public RitualMagicSpell(DataFile dataFile, RitualMagicSpell ritualMagicSpell, boolean deep, boolean forSheet) {
        super(dataFile, ritualMagicSpell, deep, forSheet);
        mPoints = forSheet ? ritualMagicSpell.mPoints : 0;
        updateLevel(false);
    }

    /**
     * Loads a ritual magic spell and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param reader   The XML reader to load from.
     * @param state    The {@link LoadState} to use.
     */
    public RitualMagicSpell(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
        this(dataFile);
        load(reader, state);
        if (!(dataFile instanceof GURPSCharacter) && !(dataFile instanceof Template)) {
            mPoints = 0;
        }
    }

    /**
     * @return The name (without specialization) of the skill to which Ritual Magic spells default.
     */
    public static String getDefaultSkillName() {
        // Using "College" as the default skill is only a convention
        return I18n.Text("College");
    }

    /**
     * Call to force an update of the level and relative level for this spell.
     *
     * @param notify Whether or not a notification should be issued on a change.
     */
    @Override
    public void updateLevel(boolean notify) {
        SkillLevel skillLevel = calculateLevel(getCharacter(), getName(), getCollege(), getPowerSource(), getCategories(), getDifficulty(), mPrerequisiteSpellsCount, mPoints);
        if (mLevel == null || !mLevel.isSameLevelAs(skillLevel)) {
            mLevel = skillLevel;
            if (notify) {
                notify(ID_LEVEL, this);
            }
        }
    }

    /**
     * Calculates the spell level.
     *
     * @param character         The character the spell will be attached to.
     * @param name              The name of the spell.
     * @param college           The college of the spell.
     * @param powerSource       The power source of the spell.
     * @param difficulty        The difficulty of the spell.
     * @param prereqSpellsCount The number of prerequisite spells for the spell with this name.
     * @param points            The number of points spent in the spell.
     * @return The calculated spell level.
     */
    public static SkillLevel calculateLevel(GURPSCharacter character, String name, String college, String powerSource, Set<String> categories, SkillDifficulty difficulty, int prereqSpellsCount, int points) {
        // Compute initial level using the technique formula
        SkillDefault def = makeSkillDefault(college, prereqSpellsCount);
        // TODO: Is the specialiaztion parameter the specialization of the technique or the full name of the default skill?
        SkillLevel skillLevel = Technique.calculateTechniqueLevel(character, name, null, categories, def, difficulty, points, true, 0);
        // FIXME: calculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
        skillLevel.mRelativeLevel += def.getModifier();

        // And then apply bonuses for spells
        if (character != null) {
            int           bonusLevels = 0;
            StringBuilder toolTip     = new StringBuilder(skillLevel.mToolTip);
            // TODO: Should college-wide bonuses be applied direcly to the College skills?
            bonusLevels += Spell.getSpellBonusesFor(character, ID_COLLEGE, college, categories, toolTip);
            bonusLevels += Spell.getSpellBonusesFor(character, ID_POWER_SOURCE, powerSource, categories, toolTip);
            bonusLevels += Spell.getSpellBonusesFor(character, ID_NAME, name, categories, toolTip);
            skillLevel.mLevel += bonusLevels;
            skillLevel.mRelativeLevel += bonusLevels;
        }
        return skillLevel;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof RitualMagicSpell) {
            RitualMagicSpell that = (RitualMagicSpell) obj;
            if (mPrerequisiteSpellsCount != that.mPrerequisiteSpellsCount) {
                return false;
            }
            return super.isEquivalentTo(obj);
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Spell (Ritual Magic)");
    }

    @Override
    public String getXMLTagName() {
        return TAG_RITUAL_MAGIC_SPELL;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Spell (Ritual Magic)");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mPrerequisiteSpellsCount = 0;
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new RitualMagicSpellEditor(this);
    }

    /**
     * @param builder The {@link StringBuilder} to append this technique's satisfied/unsatisfied
     *                description to. May be {@code null}.
     * @param prefix  The prefix to add to each line appended to the builder.
     * @return {@code true} if this technique has its default satisfied.
     */
    public boolean satisfied(StringBuilder builder, String prefix) {
        String college = getCollege();
        if (college == null || college.isBlank()) {
            if (builder != null) {
                String format = I18n.Text("{0}Must be assigned to a college\n");
                builder.append(MessageFormat.format(format, prefix));
            }
            return false; // Do not check further requirements
        }

        boolean result    = true;
        String  skillName = I18n.Text("College");
        String  skillSpec = getCollege();
        Skill   skill     = getCharacter().getBestSkillNamed(skillName, skillSpec, false, new HashSet<>());

        if (skill == null) {
            if (builder != null) {
                String format = I18n.Text("{0}Requires a skill named {1} ({2})\n");
                builder.append(MessageFormat.format(format, prefix, skillName, skillSpec));
            }
            result = false;
        }
        return result;
    }

    /**
     * Creates a SkillDefault for this spell. Intended for compatibility with other methods that
     * require a SkillDefault instance.
     *
     * @param college                 The college of the spell.
     * @param prerequisiteSpellsCount The (unsigned) number of prerequisite spells of this spell.
     * @return A {@link SkillDefault}.
     */
    public static SkillDefault makeSkillDefault(String college, int prerequisiteSpellsCount) {
        // If college is invalid, return an (hopefully) invalid default
        if (college == null || college.isBlank()) {
            return new SkillDefault(SkillDefaultType.Skill, null, college, -prerequisiteSpellsCount);
        }
        return new SkillDefault(SkillDefaultType.Skill, getDefaultSkillName(), college, -prerequisiteSpellsCount);
    }

    public int getPrerequisiteSpellsCount() {
        return mPrerequisiteSpellsCount;
    }

    public boolean setPrerequisiteSpellsCount(int prerequisiteSpellsCount) {
        if (mPrerequisiteSpellsCount != prerequisiteSpellsCount) {
            mPrerequisiteSpellsCount = prerequisiteSpellsCount;
            // TODO: Should this call something like notifySingle(ID_PREREQ_COUNT) like the setters in the Spell class do?
            updateLevel(true);
            return true;
        }
        return false;
    }
}
