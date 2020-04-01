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

/** A GURPS Spell for the Ritual Magic system. */
public class SpellTechnique extends Spell {
    private static final int          CURRENT_VERSION     = 1;
    /** The XML tag used for items. */
    public  static final String TAG_SPELL_TECHNIQUE = "spell_technique";
    /** The (positive) number of spell prerequsites needed to cast this spell. Used as skill penalty relative to the College skill */
    private              int    mSpellPrerequisiteCount;

    /**
     * Creates a new ritual magic spell.
     *
     * @param dataFile    The data file to associate it with.
     */
    public SpellTechnique(DataFile dataFile) {
        super(dataFile, false);
        mPoints  = 0;
        updateLevel(false);
    }

    /**
     * Creates a clone of an existing ritual magic spell and associates it with the specified data file.
     *
     * @param dataFile       The data file to associate it with.
     * @param spellTechnique The spell to clone.
     * @param deep           Whether or not to clone the children, grandchildren, etc.
     * @param forSheet       Whether this is for a character sheet or a list.
     */
    public SpellTechnique(DataFile dataFile, SpellTechnique spellTechnique, boolean deep, boolean forSheet) {
        super(dataFile, spellTechnique, deep, forSheet);
        mPoints = forSheet ? spellTechnique.mPoints : 0;
        updateLevel(false);
    }

    /**
     * Loads a ritual magic spell and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param reader   The XML reader to load from.
     * @param state    The {@link LoadState} to use.
     */
    public SpellTechnique(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
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
        GURPSCharacter  character      = getCharacter();
        String          name           = getName();
        String          specialization = null;
        Set<String>     categories     = getCategories();
        SkillDefault    def            = getDefault();
        SkillDifficulty difficulty     = getDifficulty();
        int             points         = mPoints;
        boolean         limited        = true;
        int             limitModifier  = 0;

        SkillLevel skillLevel = Technique.calculateTechniqueLevel(character, name, specialization, categories, def, difficulty, points, limited, limitModifier);
        if (skillLevel.mLevel != -1) {
            int bonusLevels = 0;
            StringBuilder toolTip = new StringBuilder(skillLevel.mToolTip);
            bonusLevels += getSpellBonuses(character, ID_COLLEGE,      getCollege(),     categories, toolTip);
            bonusLevels += getSpellBonuses(character, ID_POWER_SOURCE, getPowerSource(), categories, toolTip);
            bonusLevels += getSpellBonuses(character, ID_NAME,         name,             categories, toolTip);
            skillLevel.mLevel         += bonusLevels;
            skillLevel.mRelativeLevel += bonusLevels;
        }

        if(mLevel == null || !mLevel.isSameLevelAs(skillLevel)) {
            mLevel = skillLevel;
            if(notify) {
                notify(ID_LEVEL, this);
            }
        }
    }

    // Copied from Spell class
    private static int getSpellBonuses(GURPSCharacter character, String id, String qualifier, Set<String> categories, StringBuilder toolTip) {
        int level = character.getIntegerBonusFor(id, toolTip);
        level += character.getIntegerBonusFor(id + '/' + qualifier.toLowerCase(), toolTip);
        level += character.getSpellComparedIntegerBonusFor(id + '*', qualifier, categories, toolTip);
        return level;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this)
            return true;
        if (obj instanceof SpellTechnique) {
            SpellTechnique that = (SpellTechnique) obj;
            if (mSpellPrerequisiteCount != that.mSpellPrerequisiteCount)
                return false;
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
        return TAG_SPELL_TECHNIQUE;
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
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new SpellTechniqueEditor(this);
    }

    /**
     * @param builder The {@link StringBuilder} to append this technique's satisfied/unsatisfied
     *                description to. May be {@code null}.
     * @param prefix  The prefix to add to each line appended to the builder.
     * @return {@code true} if this technique has its default satisfied.
     */
    public boolean satisfied(StringBuilder builder, String prefix) {
        if(getCollege() == null || getCollege().isBlank()) {
            if (builder != null) {
                String format = I18n.Text("{0}Must be assigned to a college\n");
                builder.append(MessageFormat.format(format, prefix));
            }
            return false; // Do not check further requirements
        }

        boolean result = true;
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

    public SkillDefault getDefault() {
        return new SkillDefault(SkillDefaultType.Skill, I18n.Text("College"), getCollege(), -mSpellPrerequisiteCount);
    }

    public int getSpellPrerequisiteCount() {
        return mSpellPrerequisiteCount;
    }

    public boolean setSpellPrerequisiteCount(int spellPrerequisiteCount) {
        if (mSpellPrerequisiteCount != spellPrerequisiteCount) {
            mSpellPrerequisiteCount = spellPrerequisiteCount;
            return true;
        }
        return false;
    }
}
