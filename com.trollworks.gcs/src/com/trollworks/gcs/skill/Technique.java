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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;
import java.text.MessageFormat;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

/** A GURPS Technique. */
public class Technique extends Skill {
    /** The XML tag used for items. */
    public static final  String       TAG_TECHNIQUE   = "technique";
    private static final String       ATTRIBUTE_LIMIT = "limit";
    private static final String KEY_DEFAULT = "default";
    private              SkillDefault mDefault;
    private              boolean      mLimited;
    private              int          mLimitModifier;

    /**
     * Calculates the technique level.
     *
     * @param character      The character the technique will be attached to.
     * @param name           The name of the technique.
     * @param specialization The specialization of the technique.
     * @param def            The default the technique is based on.
     * @param difficulty     The difficulty of the technique.
     * @param points         The number of points spent in the technique.
     * @param requirePoints  Whether only skills that have points in them are considered for
     *                       defaults.
     * @param limited        Whether the technique has been limited or not.
     * @param limitModifier  The maximum bonus the technique can grant.
     * @return The calculated technique level.
     */
    public static SkillLevel calculateTechniqueLevel(GURPSCharacter character, String name, String specialization, Set<String> categories, SkillDefault def, SkillDifficulty difficulty, int points, boolean requirePoints, boolean limited, int limitModifier) {
        StringBuilder toolTip       = new StringBuilder();
        int           relativeLevel = 0;
        int           level         = Integer.MIN_VALUE;
        if (character != null) {
            level = getBaseLevel(character, def, requirePoints);
            if (level != Integer.MIN_VALUE) {
                int baseLevel = level;
                level += def.getModifier();
                if (difficulty == SkillDifficulty.H) {
                    points--;
                }
                if (points > 0) {
                    relativeLevel = points;
                }
                if (level != Integer.MIN_VALUE) {
                    relativeLevel += character.getIntegerBonusFor(ID_NAME + "/" + name.toLowerCase(), toolTip) + character.getSkillComparedIntegerBonusFor(ID_NAME + "*", name, specialization, categories, toolTip);
                    level += relativeLevel;
                }
                if (limited) {
                    int max = baseLevel + limitModifier;
                    if (level > max) {
                        relativeLevel -= level - max;
                        level = max;
                    }
                }
            }
        }
        return new SkillLevel(level, relativeLevel, toolTip);
    }

    private static int getBaseLevel(GURPSCharacter character, SkillDefault def, boolean requirePoints) {
        SkillDefaultType type = def.getType();
        if (type == SkillDefaultType.Skill) {
            Skill skill = getBaseSkill(character, def, requirePoints);
            return skill != null ? skill.getLevel() : Integer.MIN_VALUE;
        }
        // Take the modifier back out, as we wanted the base, not the final value.
        return type.getSkillLevelFast(character, def, true, null) - def.getModifier();
    }

    /**
     * Creates a string suitable for displaying the level.
     *
     * @param level         The skill level.
     * @param relativeLevel The relative skill level.
     * @param modifier      The modifer to the skill level.
     * @return The formatted string.
     */
    public static String getTechniqueDisplayLevel(int level, int relativeLevel, int modifier) {
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + Numbers.formatWithForcedSign(relativeLevel + modifier);
    }

    /**
     * Creates a new technique.
     *
     * @param dataFile The data file to associate it with.
     */
    public Technique(DataFile dataFile) {
        super(dataFile, false);
        mDefault = new SkillDefault(SkillDefaultType.Skill, I18n.Text("Skill"), null, 0);
        updateLevel(false);
    }

    /**
     * Creates a clone of an existing technique and associates it with the specified data file.
     *
     * @param dataFile  The data file to associate it with.
     * @param technique The technique to clone.
     * @param forSheet  Whether this is for a character sheet or a list.
     */
    public Technique(DataFile dataFile, Technique technique, boolean forSheet) {
        super(dataFile, technique, false, forSheet);
        if (forSheet) {
            mPoints = technique.mPoints;
        } else {
            mPoints = getDifficulty() == SkillDifficulty.A ? 1 : 2;
        }
        mDefault = new SkillDefault(technique.mDefault);
        mLimited = technique.mLimited;
        mLimitModifier = technique.mLimitModifier;
        updateLevel(false);
    }

    public Technique(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile);
        load(m, state);
        if (!(dataFile instanceof GURPSCharacter) && !(dataFile instanceof Template)) {
            mPoints = getDifficulty() == SkillDifficulty.A ? 1 : 2;
        }
    }

    /**
     * Loads a technique and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param reader   The XML reader to load from.
     * @param state    The {@link LoadState} to use.
     */
    public Technique(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
        this(dataFile);
        load(reader, state);
        if (!(dataFile instanceof GURPSCharacter) && !(dataFile instanceof Template)) {
            mPoints = getDifficulty() == SkillDifficulty.A ? 1 : 2;
        }
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Technique && super.isEquivalentTo(obj)) {
            Technique row = (Technique) obj;
            if (mLimited == row.mLimited && mLimitModifier == row.mLimitModifier) {
                return mDefault.equals(row.mDefault);
            }
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Technique");
    }

    @Override
    public String getJSONTypeName() {
        return TAG_TECHNIQUE;
    }

    @Override
    public String getXMLTagName() {
        return TAG_TECHNIQUE;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Technique");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mDefault = new SkillDefault(SkillDefaultType.Skill, I18n.Text("Skill"), null, 0);
        mLimited = false;
        mLimitModifier = 0;
    }

    @Override
    protected void loadAttributes(XMLReader reader, LoadState state) {
        String value = reader.getAttribute(ATTRIBUTE_LIMIT);
        if (value != null && !value.isEmpty()) {
            mLimited = true;
            try {
                mLimitModifier = Integer.parseInt(value);
            } catch (Exception exception) {
                mLimited = false;
                mLimitModifier = 0;
            }
        }
        super.loadAttributes(reader, state);
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        if (SkillDefault.TAG_ROOT.equals(reader.getName())) {
            mDefault = new SkillDefault(reader);
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        if (m.has(ATTRIBUTE_LIMIT)) {
            mLimited = true;
            mLimitModifier = m.getInt(ATTRIBUTE_LIMIT);
        }
        mDefault = new SkillDefault(m.getMap(KEY_DEFAULT), false);
    }

    @Override
    protected void saveSelf(JsonWriter w, boolean forUndo) throws IOException {
        super.saveSelf(w, forUndo);
        if (mLimited) {
            w.keyValue(ATTRIBUTE_LIMIT, mLimitModifier);
        }
        w.key(KEY_DEFAULT);
        mDefault.save(w, false);
    }

    /**
     * @param builder The {@link StringBuilder} to append this technique's satisfied/unsatisfied
     *                description to. May be {@code null}.
     * @param prefix  The prefix to add to each line appended to the builder.
     * @return {@code true} if this technique has its default satisfied.
     */
    public boolean satisfied(StringBuilder builder, String prefix) {
        if (mDefault.getType().isSkillBased()) {
            Skill   skill     = getCharacter().getBestSkillNamed(mDefault.getName(), mDefault.getSpecialization(), false, new HashSet<>());
            boolean satisfied = skill != null && (skill instanceof Technique || skill.getPoints() > 0);
            if (!satisfied && builder != null) {
                if (skill == null) {
                    builder.append(MessageFormat.format(I18n.Text("{0}Requires a skill named {1}\n"), prefix, mDefault.getFullName()));
                } else {
                    builder.append(MessageFormat.format(I18n.Text("{0}Requires at least 1 point in the skill named {1}\n"), prefix, mDefault.getFullName()));
                }
            }
            return satisfied;
        }
        return true;
    }

    @Override
    protected SkillLevel calculateLevelSelf() {
        return calculateTechniqueLevel(getCharacter(), getName(), getSpecialization(), getCategories(), getDefault(), getDifficulty(), getPoints(), true, isLimited(), getLimitModifier());
    }

    @Override
    public void updateLevel(boolean notify) {
        if (mDefault != null) {
            super.updateLevel(notify);
        }
    }

    /**
     * @param difficulty The difficulty to set.
     * @return Whether it was modified or not.
     */
    public boolean setDifficulty(SkillDifficulty difficulty) {
        return setDifficulty(getAttribute(), difficulty);
    }

    @Override
    public String getSpecialization() {
        return mDefault.getFullName();
    }

    @Override
    public boolean setSpecialization(String specialization) {
        return false;
    }

    @Override
    public String getTechLevel() {
        return null;
    }

    @Override
    public boolean setTechLevel(String techLevel) {
        return false;
    }

    /** @return The default to base the technique on. */
    public SkillDefault getDefault() {
        return mDefault;
    }

    /**
     * @param def The new default to base the technique on.
     * @return Whether anything was changed.
     */
    public boolean setDefault(SkillDefault def) {
        if (!mDefault.equals(def)) {
            mDefault = new SkillDefault(def);
            return true;
        }
        return false;
    }

    @Override
    public void setDifficultyFromText(String text) {
        text = text.trim();
        if (SkillDifficulty.A.name().equalsIgnoreCase(text)) {
            setDifficulty(SkillDifficulty.A);
        } else if (SkillDifficulty.H.name().equalsIgnoreCase(text)) {
            setDifficulty(SkillDifficulty.H);
        }
    }

    @Override
    public String getDifficultyAsText(boolean localized) {
        SkillDifficulty difficulty = getDifficulty();
        return localized ? difficulty.toString() : difficulty.name();
    }

    /** @return Whether the maximum level is limited. */
    public boolean isLimited() {
        return mLimited;
    }

    /**
     * Sets whether the maximum level is limited.
     *
     * @param limited The value to set.
     * @return Whether anything was changed.
     */
    public boolean setLimited(boolean limited) {
        if (limited != mLimited) {
            mLimited = limited;
            return true;
        }
        return false;
    }

    /** @return The limit modifier. */
    public int getLimitModifier() {
        return mLimitModifier;
    }

    /**
     * Sets the value of limit modifier.
     *
     * @param limitModifier The value to set.
     * @return Whether anything was changed.
     */
    public boolean setLimitModifier(int limitModifier) {
        if (mLimitModifier != limitModifier) {
            mLimitModifier = limitModifier;
            return true;
        }
        return false;
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new TechniqueEditor(this);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        mDefault.fillWithNameableKeys(set);
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        super.applyNameableKeys(map);
        mDefault.applyNameableKeys(map);
    }

    @Override
    public String getModifierNotes() {
        StringBuilder buffer = new StringBuilder(super.getModifierNotes());
        if (buffer.length() > 0) {
            buffer.append(' ');
        }
        buffer.append(I18n.Text("Default: "));
        buffer.append(mDefault);
        return buffer.toString();
    }

    @Override
    public Skill getDefaultSkill() {
        return getBaseSkill(getCharacter(), mDefault, true);
    }

    @Override
    public int swapDefault() {
        // Do nothing: Default is fixed
        return 0;
    }

    @Override
    public boolean canSwapDefaults(Skill skill) {
        return false;
    }
}
