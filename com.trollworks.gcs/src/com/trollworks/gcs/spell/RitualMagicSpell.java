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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/**
 * A GURPS Spell for the Ritual Magic system.
 * <p>
 * Ritual Magic spells are techniques and as such they must default to a skill. The convention used
 * is to default to a skill named "Ritual Magic" whose specialization is the actual college of the
 * spell.
 */
public class RitualMagicSpell extends Spell {
    public static final  String KEY_RITUAL_MAGIC_SPELL  = "ritual_magic_spell";
    private static final String KEY_BASE_SKILL_NAME     = "base_skill";
    private static final String KEY_PREREQ_COUNT        = "prereq_count";
    private static final String DEFAULT_BASE_SKILL_NAME = "Ritual Magic";

    private String mBaseSkillName;
    private int    mPrerequisiteSpellsCount;

    /**
     * Creates a new ritual magic spell.
     *
     * @param dataFile The data file to associate it with.
     */
    public RitualMagicSpell(DataFile dataFile) {
        super(dataFile, false);
        mBaseSkillName = DEFAULT_BASE_SKILL_NAME;
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
        mBaseSkillName = ritualMagicSpell.mBaseSkillName;
        mPoints = forSheet ? ritualMagicSpell.mPoints : 0;
        mPrerequisiteSpellsCount = ritualMagicSpell.mPrerequisiteSpellsCount;
        updateLevel(false);
    }

    public RitualMagicSpell(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile);
        load(dataFile, m, state);
        if (!(dataFile instanceof GURPSCharacter) && !(dataFile instanceof Template)) {
            mPoints = 0;
        }
    }

    /** @return The calculated spell skill level. */
    @Override
    protected SkillLevel calculateLevelSelf() {
        return calculateLevel(getCharacter(), getName(), getBaseSkillName(), getColleges(), getPowerSource(), getCategories(), getDifficulty(), mPrerequisiteSpellsCount, getPoints());
    }

    /**
     * Calculates the spell level.
     *
     * @param character         The character the spell will be attached to.
     * @param name              The name of the spell.
     * @param baseSkillName     The base name of the skill the Ritual Magic Spell defaults from.
     * @param colleges          The colleges of the spell.
     * @param powerSource       The power source of the spell.
     * @param difficulty        The difficulty of the spell.
     * @param prereqSpellsCount The number of prerequisite spells for the spell with this name.
     * @param points            The number of points spent in the spell.
     * @return The calculated spell level.
     */
    public static SkillLevel calculateLevel(GURPSCharacter character, String name, String baseSkillName, List<String> colleges, String powerSource, Set<String> categories, SkillDifficulty difficulty, int prereqSpellsCount, int points) {
        if (colleges == null) {
            colleges = new ArrayList<>();
        }
        SkillLevel skillLevel = null;
        if (colleges.isEmpty()) {
            skillLevel = determineSkillLevelForCollege(character, name, baseSkillName, "", categories, difficulty, prereqSpellsCount, points);
        } else {
            for (String college : colleges) {
                SkillLevel si = determineSkillLevelForCollege(character, name, baseSkillName, college, categories, difficulty, prereqSpellsCount, points);
                if (skillLevel == null || skillLevel.mLevel < si.mLevel) {
                    skillLevel = si;
                }
            }
        }
        // Apply bonuses for spells
        if (character != null) {
            StringBuilder tip         = new StringBuilder(skillLevel.mToolTip);
            int           bonusLevels = Spell.getBestCollegeSpellBonus(character, categories, colleges, tip);
            bonusLevels += Spell.getSpellBonusesFor(character, ID_POWER_SOURCE, powerSource, categories, tip);
            bonusLevels += Spell.getSpellBonusesFor(character, ID_NAME, name, categories, tip);
            skillLevel.mLevel += bonusLevels;
            skillLevel.mRelativeLevel += bonusLevels;
            skillLevel.mToolTip = tip.toString();
        }
        return skillLevel;
    }

    private static SkillLevel determineSkillLevelForCollege(GURPSCharacter character, String name, String baseSkillName, String college, Set<String> categories, SkillDifficulty difficulty, int prereqSpellsCount, int points) {
        SkillDefault def        = new SkillDefault("skill", college.isBlank() ? null : baseSkillName, college, -prereqSpellsCount);
        SkillLevel   skillLevel = Technique.calculateTechniqueLevel(character, name, college, categories, def, difficulty, points, false, true, 0);
        // calculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
        skillLevel.mRelativeLevel += def.getModifier();

        SkillDefault fallbackDef        = new SkillDefault("skill", college.isBlank() ? null : baseSkillName, null, -(6 + prereqSpellsCount));
        SkillLevel   fallbackSkillLevel = Technique.calculateTechniqueLevel(character, name, college, categories, fallbackDef, difficulty, points, false, true, 0);
        // calculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
        fallbackSkillLevel.mRelativeLevel += fallbackDef.getModifier();

        return skillLevel.mLevel >= fallbackSkillLevel.mLevel ? skillLevel : fallbackSkillLevel;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof RitualMagicSpell other) {
            if (mPrerequisiteSpellsCount != other.mPrerequisiteSpellsCount) {
                return false;
            }
            return super.isEquivalentTo(obj);
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.text("Ritual Magic Spell");
    }

    @Override
    public String getJSONTypeName() {
        return KEY_RITUAL_MAGIC_SPELL;
    }

    @Override
    public String getRowType() {
        return I18n.text("Ritual Magic Spell");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mBaseSkillName = DEFAULT_BASE_SKILL_NAME;
        mPrerequisiteSpellsCount = 0;
        mPoints = 0;
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mBaseSkillName = m.getString(KEY_BASE_SKILL_NAME);
        mPrerequisiteSpellsCount = m.getInt(KEY_PREREQ_COUNT);
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        super.saveSelf(w, saveType);
        w.keyValue(KEY_BASE_SKILL_NAME, mBaseSkillName);
        w.keyValueNot(KEY_PREREQ_COUNT, mPrerequisiteSpellsCount, 0);
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
        List<String> colleges = getColleges();
        if (colleges.isEmpty()) {
            if (builder != null) {
                builder.append(MessageFormat.format(I18n.text("{0}Must be assigned to a college\n"), prefix));
            }
            return false;
        }
        GURPSCharacter character = getCharacter();
        for (String college : colleges) {
            if (character.getBestSkillNamed(mBaseSkillName, college, false, new HashSet<>()) != null) {
                return true;
            }
        }
        if (character.getBestSkillNamed(mBaseSkillName, null, false, new HashSet<>()) == null) {
            if (builder != null) {
                builder.append(MessageFormat.format(I18n.text("{0}Requires a skill named {1} ({2})"), prefix, mBaseSkillName, colleges.get(0)));
                int size = colleges.size();
                for (int i = 1; i < size; i++) {
                    builder.append(MessageFormat.format(I18n.text("or {1} ({2})"), mBaseSkillName, colleges.get(i)));
                }
                builder.append("\n");
            }
            return false;
        }
        return true;
    }

    public String getBaseSkillName() {
        return mBaseSkillName;
    }

    public boolean setBaseSkillName(String name) {
        if (name == null) {
            name = "";
        }
        if (!mBaseSkillName.equals(name)) {
            mBaseSkillName = name;
            updateLevel(true);
            return true;
        }
        return false;
    }

    public int getPrerequisiteSpellsCount() {
        return mPrerequisiteSpellsCount;
    }

    public boolean setPrerequisiteSpellsCount(int prerequisiteSpellsCount) {
        if (mPrerequisiteSpellsCount != prerequisiteSpellsCount) {
            mPrerequisiteSpellsCount = prerequisiteSpellsCount;
            updateLevel(true);
            return true;
        }
        return false;
    }

    /** @param text The combined attribute/difficulty to set. */
    public void setDifficultyFromText(String text) {
        String input = text.trim();
        for (SkillDifficulty difficulty : SkillDifficulty.values()) {
            if (difficulty.name().equalsIgnoreCase(input)) {
                setDifficulty(getAttribute(), difficulty);
                return;
            }
        }
    }
}
