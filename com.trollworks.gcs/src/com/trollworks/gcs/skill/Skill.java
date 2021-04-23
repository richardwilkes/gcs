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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/** A GURPS Skill. */
public class Skill extends ListRow implements HasSourceReference {
    private static final int    CURRENT_JSON_VERSION    = 1;
    public static final  String KEY_SKILL               = "skill";
    public static final  String KEY_SKILL_CONTAINER     = "skill_container";
    private static final String KEY_NAME                = "name";
    private static final String KEY_SPECIALIZATION      = "specialization";
    private static final String KEY_TECH_LEVEL          = "tech_level";
    private static final String KEY_DIFFICULTY          = "difficulty";
    private static final String KEY_POINTS              = "points";
    private static final String KEY_REFERENCE           = "reference";
    private static final String KEY_ENCUMBRANCE_PENALTY = "encumbrance_penalty_multiplier";
    private static final String KEY_DEFAULTED_FROM      = "defaulted_from";
    private static final String KEY_WEAPONS             = "weapons";

    public static final String ID_NAME   = "skill.name";
    public static final String ID_POINTS = "skill.points";

    private   String            mName;
    private   String            mSpecialization;
    private   String            mTechLevel;
    private   SkillLevel        mLevel;
    private   SkillAttribute    mAttribute;
    private   SkillDifficulty   mDifficulty;
    protected int               mPoints;
    private   String            mReference;
    private   int               mEncumbrancePenaltyMultiplier;
    private   List<WeaponStats> mWeapons;
    private   SkillDefault      mDefaultedFrom;

    /**
     * Creates a string suitable for displaying the level.
     *
     * @param level         The skill level.
     * @param relativeLevel The relative skill level.
     * @param attribute     The attribute the skill is based on.
     * @param isContainer   Whether this skill is a container or not.
     * @return The formatted string.
     */
    public static String getSkillDisplayLevel(int level, int relativeLevel, SkillAttribute attribute, boolean isContainer) {
        if (isContainer) {
            return "";
        }
        if (level < 0) {
            return "-";
        }
        return Numbers.format(level) + "/" + attribute + Numbers.formatWithForcedSign(relativeLevel);
    }

    /**
     * Creates a new skill.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public Skill(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
        mName = getLocalizedName();
        mSpecialization = "";
        mTechLevel = null;
        mAttribute = SkillAttribute.DX;
        mDifficulty = SkillDifficulty.A;
        mPoints = 1;
        mReference = "";
        mWeapons = new ArrayList<>();
        updateLevel(false);
    }

    /**
     * Creates a clone of an existing skill and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param skill    The skill to clone.
     * @param deep     Whether or not to clone the children, grandchildren, etc.
     * @param forSheet Whether this is for a character sheet or a list.
     */
    public Skill(DataFile dataFile, Skill skill, boolean deep, boolean forSheet) {
        super(dataFile, skill);
        mName = skill.mName;
        mSpecialization = skill.mSpecialization;
        mTechLevel = skill.mTechLevel;
        mAttribute = skill.mAttribute;
        mDifficulty = skill.mDifficulty;
        mPoints = forSheet ? skill.mPoints : 1;
        mReference = skill.mReference;
        mEncumbrancePenaltyMultiplier = skill.mEncumbrancePenaltyMultiplier;
        if (forSheet && dataFile instanceof GURPSCharacter) {
            if (mTechLevel != null) {
                mTechLevel = ((GURPSCharacter) dataFile).getProfile().getTechLevel();
            }
        } else {
            if (mTechLevel != null && !mTechLevel.trim().isEmpty()) {
                mTechLevel = "";
            }
        }
        mWeapons = new ArrayList<>(skill.mWeapons.size());
        for (WeaponStats weapon : skill.mWeapons) {
            if (weapon instanceof MeleeWeaponStats) {
                mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
            } else if (weapon instanceof RangedWeaponStats) {
                mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
            }
        }
        updateLevel(false);
        if (deep) {
            int count = skill.getChildCount();
            for (int i = 0; i < count; i++) {
                Row row = skill.getChild(i);
                if (row instanceof Technique) {
                    addChild(new Technique(dataFile, (Technique) row, forSheet));
                } else {
                    addChild(new Skill(dataFile, (Skill) row, true, forSheet));
                }
            }
        }
    }

    public Skill(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, m.getString(DataFile.KEY_TYPE).equals(KEY_SKILL_CONTAINER));
        load(m, state);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Skill && getClass() == obj.getClass() && super.isEquivalentTo(obj)) {
            Skill row = (Skill) obj;
            if (mLevel.isSameLevelAs(row.mLevel)) {
                if (mPoints == row.mPoints) {
                    if (mEncumbrancePenaltyMultiplier == row.mEncumbrancePenaltyMultiplier) {
                        if (mAttribute == row.mAttribute) {
                            if (mDifficulty == row.mDifficulty) {
                                if (mName.equals(row.mName)) {
                                    if (Objects.equals(mTechLevel, row.mTechLevel)) {
                                        if (mSpecialization.equals(row.mSpecialization)) {
                                            if (mReference.equals(row.mReference)) {
                                                return mWeapons.equals(row.mWeapons);
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Skill");
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_SKILL_CONTAINER : KEY_SKILL;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Skill");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mName = getLocalizedName();
        mSpecialization = "";
        mTechLevel = null;
        mAttribute = SkillAttribute.DX;
        mDifficulty = SkillDifficulty.A;
        mPoints = 1;
        mReference = "";
        mEncumbrancePenaltyMultiplier = 0;
        mWeapons = new ArrayList<>();
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        mName = m.getString(KEY_NAME);
        mReference = m.getString(KEY_REFERENCE);
        if (!canHaveChildren()) {
            mSpecialization = m.getString(KEY_SPECIALIZATION);
            if (m.has(KEY_TECH_LEVEL)) {
                mTechLevel = m.getString(KEY_TECH_LEVEL);
                if (!mTechLevel.isBlank() && getDataFile() instanceof ListFile) {
                    mTechLevel = "";
                }
            }
            mEncumbrancePenaltyMultiplier = m.getInt(KEY_ENCUMBRANCE_PENALTY);
            setDifficultyFromText(m.getString(KEY_DIFFICULTY));
            mPoints = m.getInt(KEY_POINTS);
            if (m.has(KEY_DEFAULTED_FROM)) {
                mDefaultedFrom = new SkillDefault(m.getMap(KEY_DEFAULTED_FROM), true);
            }
            if (m.has(KEY_WEAPONS)) {
                WeaponStats.loadFromJSONArray(this, m.getArray(KEY_WEAPONS), mWeapons);
            }
        }
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (KEY_SKILL.equals(type) || KEY_SKILL_CONTAINER.equals(type)) {
                addChild(new Skill(mDataFile, m, state));
            } else if (Technique.KEY_TECHNIQUE.equals(type)) {
                addChild(new Technique(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        w.keyValue(KEY_NAME, mName);
        w.keyValueNot(KEY_REFERENCE, mReference, "");
        if (!canHaveChildren()) {
            w.keyValueNot(KEY_SPECIALIZATION, mSpecialization, "");
            if (mTechLevel != null) {
                if (getCharacter() != null) {
                    w.keyValueNot(KEY_TECH_LEVEL, mTechLevel, "");
                } else {
                    w.keyValue(KEY_TECH_LEVEL, "");
                }
            }
            w.keyValueNot(KEY_ENCUMBRANCE_PENALTY, mEncumbrancePenaltyMultiplier, 0);
            w.keyValue(KEY_DIFFICULTY, getDifficultyAsText(false));
            w.keyValue(KEY_POINTS, mPoints);
            if (mDefaultedFrom != null) {
                w.key(KEY_DEFAULTED_FROM);
                mDefaultedFrom.save(w, true);
            }
            WeaponStats.saveList(w, KEY_WEAPONS, mWeapons);
        }
    }

    /** @return The weapon list. */
    public List<WeaponStats> getWeapons() {
        return Collections.unmodifiableList(mWeapons);
    }

    /**
     * @param weapons The weapons to set.
     * @return Whether it was modified.
     */
    public boolean setWeapons(List<WeaponStats> weapons) {
        if (!mWeapons.equals(weapons)) {
            mWeapons = new ArrayList<>(weapons);
            for (WeaponStats weapon : mWeapons) {
                weapon.setOwner(this);
            }
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The level. */
    public int getLevel() {
        return mLevel.getLevel();
    }

    /** @return The relative level. */
    public int getRelativeLevel() {
        return mLevel.getRelativeLevel();
    }

    /** @return The tooltTip to describe how the level was calculated */
    public String getLevelToolTip() {
        return mLevel.getToolTip();
    }

    /** @return The name. */
    public String getName() {
        return mName;
    }

    /**
     * @param name The name to set.
     * @return Whether it was changed.
     */
    public boolean setName(String name) {
        if (!mName.equals(name)) {
            mName = name;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The specialization. */
    public String getSpecialization() {
        return mSpecialization;
    }

    /**
     * @param specialization The specialization to set.
     * @return Whether it was changed.
     */
    public boolean setSpecialization(String specialization) {
        if (!mSpecialization.equals(specialization)) {
            mSpecialization = specialization;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The tech level. */
    public String getTechLevel() {
        return mTechLevel;
    }

    /**
     * @param techLevel The tech level to set.
     * @return Whether it was changed.
     */
    public boolean setTechLevel(String techLevel) {
        if (!Objects.equals(mTechLevel, techLevel)) {
            mTechLevel = techLevel;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The tooltTip to describe how the points were calculated */
    public String getPointsToolTip() {
        if (canHaveChildren()) {
            return I18n.Text("The sum of the points spent by children of this container");
        }
        GURPSCharacter character = getCharacter();
        if (character != null) {
            StringBuilder tooltip = new StringBuilder();
            String        name    = getName();
            character.getSkillPointComparedIntegerBonusFor(ID_POINTS + "*", name, getSpecialization(), getCategories(), tooltip);
            character.getIntegerBonusFor(ID_POINTS + "/" + name.toLowerCase(), tooltip);
            if (!tooltip.isEmpty()) {
                return I18n.Text("Includes modifiers from") + tooltip;
            }
        }
        return "";
    }

    /** @return The points. */
    public int getPoints() {
        if (canHaveChildren()) {
            int sum = 0;
            for (Row row : getChildren()) {
                if (row instanceof Skill) {
                    sum += ((Skill) row).getPoints();
                }
            }
            return sum;
        }
        int            points    = mPoints;
        GURPSCharacter character = getCharacter();
        if (character != null) {
            String name = getName();
            points += character.getSkillPointComparedIntegerBonusFor(ID_POINTS + "*", name, getSpecialization(), getCategories());
            points += character.getIntegerBonusFor(ID_POINTS + "/" + name.toLowerCase());
            if (points < 0) {
                points = 0;
            }
        }
        return points;
    }

    /** @return the unmodified points */
    public int getRawPoints() {
        return canHaveChildren() ? 0 : mPoints;
    }

    /**
     * @param points The points to set.
     * @return Whether it was changed.
     */
    public boolean setRawPoints(int points) {
        if (mPoints != points) {
            mPoints = points;
            updateLevel(false);
            notifyOfChange();
            return true;
        }
        return false;
    }

    /**
     * Call to force an update of the level and relative level for this skill or technique.
     *
     * @param notify Whether or not a notification should be issued on a change.
     */
    public void updateLevel(boolean notify) {
        SkillLevel savedLevel = mLevel;
        mLevel = calculateLevelSelf();
        if (notify) {
            if (savedLevel.isDifferentLevelThan(mLevel) || savedLevel.isDifferentRelativeLevelThan(mLevel)) {
                notifyOfChange();
            }
        }
    }

    /** @return The calculated skill level. */
    protected SkillLevel calculateLevelSelf() {
        mDefaultedFrom = getBestDefaultWithPoints();
        return calculateLevel(getCharacter(), getName(), getSpecialization(), getCategories(), getDefaults(), getAttribute(), getDifficulty(), getPoints(), new HashSet<>(), getEncumbrancePenaltyMultiplier());
    }

    /**
     * @param excludes Skills to exclude, other than this one.
     * @return The calculated level.
     */
    public int getLevel(Set<String> excludes) {
        return calculateLevel(getCharacter(), getName(), getSpecialization(), getCategories(), getDefaults(), getAttribute(), getDifficulty(), getPoints(), excludes, getEncumbrancePenaltyMultiplier()).mLevel;
    }

    /** @return The attribute. */
    public SkillAttribute getAttribute() {
        return mAttribute;
    }

    /** @return The difficulty. */
    public SkillDifficulty getDifficulty() {
        return mDifficulty;
    }

    /**
     * @param attribute  The attribute to set.
     * @param difficulty The difficulty to set.
     * @return Whether it was changed.
     */
    public boolean setDifficulty(SkillAttribute attribute, SkillDifficulty difficulty) {
        if (mAttribute != attribute || mDifficulty != difficulty) {
            mAttribute = attribute;
            mDifficulty = difficulty;
            updateLevel(false);
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The encumbrance penalty multiplier. */
    public int getEncumbrancePenaltyMultiplier() {
        return mEncumbrancePenaltyMultiplier;
    }

    /**
     * @param multiplier The multiplier to set.
     * @return Whether it was changed.
     */
    public boolean setEncumbrancePenaltyMultiplier(int multiplier) {
        multiplier = Math.min(Math.max(multiplier, 0), 9);
        if (mEncumbrancePenaltyMultiplier != multiplier) {
            mEncumbrancePenaltyMultiplier = multiplier;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getReference() {
        return mReference;
    }

    @Override
    public boolean setReference(String reference) {
        if (!mReference.equals(reference)) {
            mReference = reference;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getReferenceHighlight() {
        return getName();
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getName().toLowerCase().contains(text)) {
            return true;
        }
        if (getSpecialization().toLowerCase().contains(text)) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    @Override
    public Object getData(Column column) {
        return SkillColumn.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return SkillColumn.values()[column.getID()].getDataAsText(this);
    }

    /** @param text The combined attribute/difficulty to set. */
    public void setDifficultyFromText(String text) {
        SkillAttribute[]  attribute  = SkillAttribute.values();
        SkillDifficulty[] difficulty = SkillDifficulty.values();
        String            input      = text.trim();

        for (SkillAttribute element : attribute) {
            // We have to go backwards through the list to avoid the
            // regex grabbing the "H" in "VH".
            for (int j = difficulty.length - 1; j >= 0; j--) {
                if (input.matches("(?i).*" + element.name() + ".*/.*" + difficulty[j].name() + ".*")) {
                    setDifficulty(element, difficulty[j]);
                    return;
                }
            }
        }
    }

    /** @return The formatted attribute/difficulty. */
    public String getDifficultyAsText() {
        return getDifficultyAsText(true);
    }

    /**
     * @param localized Whether to use localized versions of attribute and difficulty.
     * @return The formatted attribute/difficulty.
     */
    public String getDifficultyAsText(boolean localized) {
        if (canHaveChildren()) {
            return "";
        }
        return (localized ? mAttribute.toString() : mAttribute.name()) + "/" + (localized ? mDifficulty.toString() : mDifficulty.name());
    }

    @Override
    public String toString() {
        StringBuilder builder = new StringBuilder();

        builder.append(getName());
        if (!canHaveChildren()) {
            String techLevel      = getTechLevel();
            String specialization = getSpecialization();

            if (techLevel != null) {
                builder.append("/TL");
                if (!techLevel.isEmpty()) {
                    builder.append(techLevel);
                }
            }

            if (!specialization.isEmpty()) {
                builder.append(" (");
                builder.append(specialization);
                builder.append(')');
            }
        }
        return builder.toString();
    }

    @Override
    public String getModifierNotes() {
        StringBuilder buffer = new StringBuilder(super.getModifierNotes());
        Skill         skill  = getDefaultSkill();
        if (skill != null && mDefaultedFrom != null) {
            if (!buffer.isEmpty()) {
                buffer.append(' ');
            }
            buffer.append(I18n.Text("Default: "));
            buffer.append(skill);
            buffer.append(mDefaultedFrom.getModifierAsString());
        }

        return buffer.toString();
    }

    @Override
    public RetinaIcon getIcon(boolean marker) {
        return marker ? Images.SKL_MARKER : Images.SKL_FILE;
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new SkillEditor(this);
    }

    /**
     * Calculates the skill level.
     *
     * @param character      The character the skill will be attached to.
     * @param name           The name of the skill.
     * @param specialization The specialization of the skill.
     * @param defaults       The defaults the skill has.
     * @param attribute      The attribute the skill is based on.
     * @param difficulty     The difficulty of the skill.
     * @param points         The number of points spent in the skill.
     * @param excludes       The set of skills to exclude from any default calculations.
     * @param encPenaltyMult The encumbrance penalty multiplier.
     * @return The calculated skill level.
     */
    public SkillLevel calculateLevel(GURPSCharacter character, String name, String specialization, Set<String> categories, List<SkillDefault> defaults, SkillAttribute attribute, SkillDifficulty difficulty, int points, Set<String> excludes, int encPenaltyMult) {
        StringBuilder toolTip       = new StringBuilder();
        int           relativeLevel = difficulty.getBaseRelativeLevel();
        int           level         = attribute.getBaseSkillLevel(character);
        if (level != Integer.MIN_VALUE) {
            if (difficulty == SkillDifficulty.W) {
                points /= 3;
            } else {
                if (mDefaultedFrom != null && mDefaultedFrom.getPoints() > 0) {
                    points += mDefaultedFrom.getPoints();
                }
            }

            if (points > 0) {
                relativeLevel = calculateRelativeLevel(points, relativeLevel);
            } else if (mDefaultedFrom != null && mDefaultedFrom.getPoints() < 0) {
                relativeLevel = mDefaultedFrom.getAdjLevel() - level;
            } else {
                level = Integer.MIN_VALUE;
                relativeLevel = 0;
            }

            if (level != Integer.MIN_VALUE) {
                level += relativeLevel;
                if (mDefaultedFrom != null) {
                    if (level < mDefaultedFrom.getAdjLevel()) {
                        level = mDefaultedFrom.getAdjLevel();
                    }
                }
                if (character != null) {
                    int bonus = character.getSkillComparedIntegerBonusFor(ID_NAME + "*", name, specialization, categories, toolTip);
                    level += bonus;
                    relativeLevel += bonus;
                    bonus = character.getIntegerBonusFor(ID_NAME + "/" + name.toLowerCase(), toolTip);
                    level += bonus;
                    relativeLevel += bonus;
                    bonus = character.getEncumbranceLevel(true).getEncumbrancePenalty() * encPenaltyMult;
                    level += bonus;
                    if (bonus != 0) {
                        toolTip.append(String.format(I18n.Text("\nEncumbrance [%d]"), Integer.valueOf(bonus)));
                    }
                }
            }
        }
        return new SkillLevel(level, relativeLevel, toolTip);
    }

    /** Swaps the default to an alternate, if possible. */
    public void swapDefault() {
        Skill baseSkill = getDefaultSkill();
        if (baseSkill != null) {
            mDefaultedFrom = getBestDefaultWithPoints(mDefaultedFrom);
            baseSkill.updateLevel(false);
            updateLevel(false);
            notifyOfChange();
        }
    }

    /**
     * Returns {@code true} if default can be swapped with {@code skill}.
     *
     * @param skill Skill to check.
     * @return {@code true} if default can be swapped with {@code skill}.
     */
    public boolean canSwapDefaults(Skill skill) {
        boolean result = false;
        if (mDefaultedFrom != null && getPoints() > 0) {
            if (skill != null && skill.hasDefaultTo(this)) {
                result = true;
            }
        }
        return result;
    }

    private boolean hasDefaultTo(Skill skill) {
        boolean result = false;
        for (SkillDefault skillDefault : getDefaults()) {
            boolean skillBased            = skillDefault.getType().isSkillBased();
            boolean nameMatches           = skillDefault.getName().equals(skill.getName());
            boolean specializationMatches = skillDefault.getSpecialization() == null || skillDefault.getSpecialization().isEmpty() || skillDefault.getSpecialization().equals(skill.getSpecialization());
            if (skillBased && nameMatches && specializationMatches) {
                result = true;
                break;
            }
        }
        return result;
    }

    private static int calculateRelativeLevel(int points, int relativeLevel) {
        if (points == 1) {
            // relativeLevel is preset to this point value
        } else if (points < 4) {
            relativeLevel++;
        } else {
            relativeLevel += 1 + points / 4;
        }
        return relativeLevel;
    }

    private SkillDefault getBestDefaultWithPoints() {
        return getBestDefaultWithPoints(null);
    }

    private SkillDefault getBestDefaultWithPoints(SkillDefault excludedDefault) {
        SkillDefault best = getBestDefault(excludedDefault);
        if (best != null) {
            GURPSCharacter character = getCharacter();
            int            baseLine  = getAttribute().getBaseSkillLevel(character) + getDifficulty().getBaseRelativeLevel();
            int            level     = best.getLevel();
            best.setAdjLevel(level);
            if (level == baseLine) {
                best.setPoints(1);
            } else if (level == baseLine + 1) {
                best.setPoints(2);
            } else if (level > baseLine + 1) {
                best.setPoints(4 * (level - (baseLine + 1)));
            } else {
                level = best.getLevel();
                if (level < 0) {
                    level = 0;
                }
                best.setPoints(-level);
            }
        }
        return best;
    }

    private SkillDefault getBestDefault(SkillDefault excludedDefault) {
        GURPSCharacter character = getCharacter();
        if (character != null) {
            Collection<SkillDefault> defaults = getDefaults();
            if (!defaults.isEmpty()) {
                int          best      = Integer.MIN_VALUE;
                SkillDefault bestSkill = null;
                String       exclude   = toString();
                Set<String>  excludes  = new HashSet<>();
                excludes.add(exclude);
                for (SkillDefault skillDefault : defaults) {
                    // For skill-based defaults, prune out any that already use a default that we
                    // are involved with
                    if (!skillDefault.equals(excludedDefault) && !isInDefaultChain(this, skillDefault, new HashSet<>())) {
                        int level = skillDefault.getType().getSkillLevel(character, skillDefault, true, excludes, !(this instanceof Technique));
                        if (skillDefault.getType().isSkillBased()) {
                            String name  = skillDefault.getName();
                            Skill  skill = character.getBestSkillNamed(name, skillDefault.getSpecialization(), true, excludes);
                            if (skill != null) {
                                level -= character.getSkillComparedIntegerBonusFor(ID_NAME + "*", name, skillDefault.getSpecialization(), skill.getCategories());
                                level -= character.getIntegerBonusFor(ID_NAME + "/" + name.toLowerCase());
                            }
                        }
                        if (level > best) {
                            best = level;
                            bestSkill = new SkillDefault(skillDefault);
                            bestSkill.setLevel(level);
                        }
                    }
                }
                return bestSkill;
            }
        }
        return null;
    }

    private boolean isInDefaultChain(Skill skill, SkillDefault skillDefault, Set<Skill> lookedAt) {
        GURPSCharacter character = getCharacter();
        if (character != null && skillDefault != null && skillDefault.getType().isSkillBased()) {
            boolean hadOne = false;
            for (Skill one : character.getSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), true, null)) {
                if (one == skill) {
                    return true;
                }
                if (lookedAt.add(one)) {
                    if (isInDefaultChain(skill, one.mDefaultedFrom, lookedAt)) {
                        return true;
                    }
                }
                hadOne = true;
            }
            return !hadOne;
        }
        return false;
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        extractNameables(set, mName);
        extractNameables(set, mSpecialization);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.fillWithNameableKeys(set);
            }
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        super.applyNameableKeys(map);
        mName = nameNameables(map, mName);
        mSpecialization = nameNameables(map, mSpecialization);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.applyNameableKeys(map);
            }
        }
    }

    /**
     * Returns the skill defaulted to.
     *
     * @param character    Character
     * @param skillDefault Skill default
     * @return Returns the skill defaulted to.
     */
    protected static Skill getBaseSkill(GURPSCharacter character, SkillDefault skillDefault, boolean requirePoints) {
        if (character != null && skillDefault != null && skillDefault.getType().isSkillBased()) {
            return character.getBestSkillNamed(skillDefault.getName(), skillDefault.getSpecialization(), requirePoints, new HashSet<>());
        }
        return null;
    }

    /**
     * Skill the skill currently Defaults to.
     *
     * @return Skill the skill currently Defaults to.
     */
    public Skill getDefaultSkill() {
        return getBaseSkill(getCharacter(), mDefaultedFrom, true);
    }

    @Override
    public String getToolTip(Column column) {
        return SkillColumn.values()[column.getID()].getToolTip(this);
    }
}
