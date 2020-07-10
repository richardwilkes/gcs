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

package com.trollworks.gcs.spell;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashSet;
import java.util.Set;

/**
 * A GURPS Spell for the Ritual Magic system.
 * <p>
 * Ritual Magic spells are techniques and as such they must default to a skill. The convention used
 * is to default to a skill named "Ritual Magic" whose specialization is the actual college of the
 * spell.
 */
public class RitualMagicSpell extends Spell {
    private static final int    CURRENT_VERSION         = 1;
    /** The XML tag used for items. */
    public static final  String TAG_RITUAL_MAGIC_SPELL  = "ritual_magic_spell";
    private static final String TAG_BASE_SKILL_NAME     = "base_skill";
    private static final String TAG_PREREQ_COUNT        = "prereq_count";
    /** The default base name of the skill Ritual Magic Spells default from. */
    public static final  String DEFAULT_BASE_SKILL_NAME = "Ritual Magic";
    private              String mBaseSkillName;
    /**
     * The (positive) number of spell prerequisites needed to cast this spell. Used as skill penalty
     * relative to the Ritual Magic skill.
     */
    private              int    mPrerequisiteSpellsCount;

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
        load(m, state);
        if (!(dataFile instanceof GURPSCharacter) && !(dataFile instanceof Template)) {
            mPoints = 0;
        }
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
     * Call to force an update of the level and relative level for this spell.
     *
     * @param notify Whether or not a notification should be issued on a change.
     */
    @Override
    public void updateLevel(boolean notify) {
        SkillLevel skillLevel = calculateLevel(getCharacter(), getName(), getBaseSkillName(), getCollege(), getPowerSource(), getCategories(), getDifficulty(), mPrerequisiteSpellsCount, mPoints);
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
     * @param baseSkillName     The base name of the skill the Ritual Magic Spell defaults from.
     * @param college           The college of the spell.
     * @param powerSource       The power source of the spell.
     * @param difficulty        The difficulty of the spell.
     * @param prereqSpellsCount The number of prerequisite spells for the spell with this name.
     * @param points            The number of points spent in the spell.
     * @return The calculated spell level.
     */
    public static SkillLevel calculateLevel(GURPSCharacter character, String name, String baseSkillName, String college, String powerSource, Set<String> categories, SkillDifficulty difficulty, int prereqSpellsCount, int points) {
        if (college == null) {
            college = "";
        }

        SkillDefault def        = new SkillDefault(SkillDefaultType.Skill, college.isBlank() ? null : baseSkillName, college, -prereqSpellsCount);
        SkillLevel   skillLevel = Technique.calculateTechniqueLevel(character, name, college, categories, def, difficulty, points, false, true, 0);
        // calculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
        skillLevel.mRelativeLevel += def.getModifier();

        SkillDefault def2        = new SkillDefault(SkillDefaultType.Skill, college.isBlank() ? null : baseSkillName, null, -(6 + prereqSpellsCount));
        SkillLevel   skillLevel2 = Technique.calculateTechniqueLevel(character, name, college, categories, def2, difficulty, points, false, true, 0);
        // calculateTechniqueLevel() does not add the default skill modifier to the relative level, only to the final level
        skillLevel2.mRelativeLevel += def2.getModifier();

        if (skillLevel.mLevel < skillLevel2.mLevel) {
            skillLevel = skillLevel2;
        }

        // And then apply bonuses for spells
        if (character != null) {
            StringBuilder tip         = new StringBuilder(skillLevel.mToolTip);
            int           bonusLevels = Spell.getSpellBonusesFor(character, ID_COLLEGE, college, categories, tip);
            bonusLevels += Spell.getSpellBonusesFor(character, ID_POWER_SOURCE, powerSource, categories, tip);
            bonusLevels += Spell.getSpellBonusesFor(character, ID_NAME, name, categories, tip);
            skillLevel.mLevel += bonusLevels;
            skillLevel.mRelativeLevel += bonusLevels;
            skillLevel.mToolTip = tip.toString();
        }
        return skillLevel;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof RitualMagicSpell) {
            RitualMagicSpell other = (RitualMagicSpell) obj;
            if (mPrerequisiteSpellsCount != other.mPrerequisiteSpellsCount) {
                return false;
            }
            return super.isEquivalentTo(obj);
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Ritual Magic Spell");
    }

    @Override
    public String getJSONTypeName() {
        return TAG_RITUAL_MAGIC_SPELL;
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
        return I18n.Text("Ritual Magic Spell");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mBaseSkillName = DEFAULT_BASE_SKILL_NAME;
        mPrerequisiteSpellsCount = 0;
        mPoints = 0;
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (TAG_BASE_SKILL_NAME.equals(name)) {
            mBaseSkillName = reader.readText().replace("\n", " ");
        } else if (TAG_PREREQ_COUNT.equals(name)) {
            mPrerequisiteSpellsCount = reader.readInteger(0);
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        mBaseSkillName = m.getString(TAG_BASE_SKILL_NAME);
        mPrerequisiteSpellsCount = m.getInt(TAG_PREREQ_COUNT);
    }

    @Override
    protected void saveSelf(JsonWriter w, boolean forUndo) throws IOException {
        super.saveSelf(w, forUndo);
        w.keyValue(TAG_BASE_SKILL_NAME, mBaseSkillName);
        w.keyValueNot(TAG_PREREQ_COUNT, mPrerequisiteSpellsCount, 0);
        // Spells assume a default of 1 point, while RM assumes a default of 0, so we have to make
        // sure it gets written
        if (mPoints == 1) {
            w.keyValue(TAG_POINTS, 1);
        }
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
                builder.append(MessageFormat.format(I18n.Text("{0}Must be assigned to a college\n"), prefix));
            }
            return false;
        }

        Skill skill = getCharacter().getBestSkillNamed(mBaseSkillName, college, false, new HashSet<>());
        if (skill == null) {
            skill = getCharacter().getBestSkillNamed(mBaseSkillName, null, false, new HashSet<>());
            if (skill == null) {
                if (builder != null) {
                    builder.append(MessageFormat.format(I18n.Text("{0}Requires a skill named {1} ({2})\n"), prefix, mBaseSkillName, college));
                }
                return false;
            }
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
}
