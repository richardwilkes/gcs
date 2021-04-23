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
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.skill.SkillAttribute;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
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
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

/** A GURPS Spell. */
public class Spell extends ListRow implements HasSourceReference {
    private static final   int    CURRENT_JSON_VERSION = 1;
    public static final    String KEY_SPELL            = "spell";
    public static final    String KEY_SPELL_CONTAINER  = "spell_container";
    private static final   String KEY_NAME             = "name";
    private static final   String KEY_TECH_LEVEL       = "tech_level";
    private static final   String KEY_COLLEGE          = "college";
    private static final   String KEY_POWER_SOURCE     = "power_source";
    private static final   String KEY_SPELL_CLASS      = "spell_class";
    private static final   String KEY_RESIST           = "resist";
    private static final   String KEY_CASTING_COST     = "casting_cost";
    private static final   String KEY_MAINTENANCE_COST = "maintenance_cost";
    private static final   String KEY_CASTING_TIME     = "casting_time";
    private static final   String KEY_DURATION         = "duration";
    protected static final String KEY_POINTS           = "points";
    private static final   String KEY_REFERENCE        = "reference";
    private static final   String KEY_ATTRIBUTE        = "attribute";
    private static final   String KEY_DIFFICULTY       = "difficulty";
    private static final   String KEY_WEAPONS          = "weapons";

    public static final String ID_NAME                = "spell.name";
    public static final String ID_COLLEGE             = "spell.college";
    public static final String ID_POWER_SOURCE        = "spell.power_source";
    public static final String ID_POINTS              = "spell.points";
    public static final String ID_POINTS_COLLEGE      = "spell.college.points";
    public static final String ID_POINTS_POWER_SOURCE = "spell.power_source.points";

    private   String            mName;
    private   String            mTechLevel;
    private   String            mCollege;
    private   String            mPowerSource;
    private   String            mSpellClass;
    private   String            mResist;
    private   String            mCastingCost;
    private   String            mMaintenance;
    private   String            mCastingTime;
    private   String            mDuration;
    protected int               mPoints;
    protected SkillLevel        mLevel;
    private   SkillAttribute    mAttribute;
    private   String            mReference;
    private   SkillDifficulty   mDifficulty;
    private   List<WeaponStats> mWeapons;

    /**
     * Creates a new spell.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public Spell(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
        mName = I18n.Text("Spell");
        mAttribute = SkillAttribute.IQ;
        mTechLevel = null;
        mCollege = "";
        mPowerSource = isContainer ? "" : getDefaultPowerSource();
        mSpellClass = isContainer ? "" : getDefaultSpellClass();
        mResist = "";
        mCastingCost = isContainer ? "" : getDefaultCastingCost();
        mMaintenance = "";
        mCastingTime = isContainer ? "" : getDefaultCastingTime();
        mDuration = isContainer ? "" : getDefaultDuration();
        mPoints = 1;
        mReference = "";
        mDifficulty = SkillDifficulty.H;
        mWeapons = new ArrayList<>();
        updateLevel(false);
    }

    /**
     * Creates a clone of an existing spell and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param spell    The spell to clone.
     * @param deep     Whether or not to clone the children, grandchildren, etc.
     * @param forSheet Whether this is for a character sheet or a list.
     */
    public Spell(DataFile dataFile, Spell spell, boolean deep, boolean forSheet) {
        super(dataFile, spell);
        mName = spell.mName;
        mAttribute = spell.mAttribute;
        mTechLevel = spell.mTechLevel;
        mCollege = spell.mCollege;
        mPowerSource = spell.mPowerSource;
        mSpellClass = spell.mSpellClass;
        mResist = spell.mResist;
        mCastingCost = spell.mCastingCost;
        mMaintenance = spell.mMaintenance;
        mCastingTime = spell.mCastingTime;
        mDuration = spell.mDuration;
        mPoints = forSheet ? spell.mPoints : 1;
        mReference = spell.mReference;
        mDifficulty = spell.mDifficulty;
        if (forSheet && dataFile instanceof GURPSCharacter) {
            if (mTechLevel != null) {
                mTechLevel = ((GURPSCharacter) dataFile).getProfile().getTechLevel();
            }
        } else {
            if (mTechLevel != null && !mTechLevel.trim().isEmpty()) {
                mTechLevel = "";
            }
        }
        mWeapons = new ArrayList<>(spell.mWeapons.size());
        for (WeaponStats weapon : spell.mWeapons) {
            if (weapon instanceof MeleeWeaponStats) {
                mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
            } else if (weapon instanceof RangedWeaponStats) {
                mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
            }
        }
        updateLevel(false);
        if (deep) {
            int count = spell.getChildCount();
            for (int i = 0; i < count; i++) {
                Row child = spell.getChild(i);
                if (child instanceof RitualMagicSpell) {
                    child = new RitualMagicSpell(dataFile, (RitualMagicSpell) child, true, forSheet);
                } else {
                    child = new Spell(dataFile, (Spell) child, true, forSheet);
                }
                addChild(child);
            }
        }
    }

    public Spell(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, m.getString(DataFile.KEY_TYPE).equals(KEY_SPELL_CONTAINER));
        load(m, state);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Spell && super.isEquivalentTo(obj)) {
            Spell row = (Spell) obj;
            if (mDifficulty == row.mDifficulty && mPoints == row.mPoints && mLevel.isSameLevelAs(row.mLevel) && mAttribute == row.mAttribute) {
                if (Objects.equals(mTechLevel, row.mTechLevel)) {
                    if (mName.equals(row.mName) && mCollege.equals(row.mCollege) && mPowerSource.equals(row.mPowerSource) && mSpellClass.equals(row.mSpellClass) && mResist.equals(row.mResist) && mReference.equals(row.mReference)) {
                        if (mCastingCost.equals(row.mCastingCost) && mMaintenance.equals(row.mMaintenance) && mCastingTime.equals(row.mCastingTime) && mDuration.equals(row.mDuration)) {
                            return mWeapons.equals(row.mWeapons);
                        }
                    }
                }
            }
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Spell");
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_SPELL_CONTAINER : KEY_SPELL;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Spell");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        boolean isContainer = canHaveChildren();
        super.prepareForLoad(state);
        mName = I18n.Text("Spell");
        mAttribute = SkillAttribute.IQ;
        mTechLevel = null;
        mCollege = "";
        mPowerSource = isContainer ? "" : getDefaultPowerSource();
        mSpellClass = isContainer ? "" : getDefaultSpellClass();
        mResist = "";
        mCastingCost = isContainer ? "" : getDefaultCastingCost();
        mMaintenance = "";
        mCastingTime = isContainer ? "" : getDefaultCastingTime();
        mDuration = isContainer ? "" : getDefaultDuration();
        mPoints = 1;
        mReference = "";
        mDifficulty = SkillDifficulty.H;
        mWeapons = new ArrayList<>();
    }

    @Override
    protected void finishedLoading(LoadState state) {
        updateLevel(false);
        super.finishedLoading(state);
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        mName = m.getString(KEY_NAME);
        mReference = m.getString(KEY_REFERENCE);
        if (!canHaveChildren()) {
            setDifficultyFromText(m.getString(KEY_DIFFICULTY));
            if (m.has(KEY_TECH_LEVEL)) {
                mTechLevel = m.getString(KEY_TECH_LEVEL);
                if (!mTechLevel.isBlank() && getDataFile() instanceof ListFile) {
                    mTechLevel = "";
                }
            }
            mAttribute = Enums.extract(m.getString(KEY_ATTRIBUTE), SkillAttribute.values(), SkillAttribute.IQ);
            mCollege = m.getString(KEY_COLLEGE);
            mPowerSource = m.getString(KEY_POWER_SOURCE);
            mSpellClass = m.getString(KEY_SPELL_CLASS);
            mResist = m.getString(KEY_RESIST);
            mCastingCost = m.getString(KEY_CASTING_COST);
            mMaintenance = m.getString(KEY_MAINTENANCE_COST);
            mCastingTime = m.getString(KEY_CASTING_TIME);
            mDuration = m.getString(KEY_DURATION);
            mPoints = m.getIntWithDefault(KEY_POINTS, 1);
            if (m.has(KEY_WEAPONS)) {
                WeaponStats.loadFromJSONArray(this, m.getArray(KEY_WEAPONS), mWeapons);
            }
        }
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (KEY_SPELL.equals(type) || KEY_SPELL_CONTAINER.equals(type)) {
                addChild(new Spell(mDataFile, m, state));
            } else if (RitualMagicSpell.KEY_RITUAL_MAGIC_SPELL.equals(type)) {
                addChild(new RitualMagicSpell(mDataFile, m, state));
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
            w.keyValue(KEY_DIFFICULTY, getDifficultyAsText(false));
            if (mTechLevel != null) {
                if (getCharacter() != null) {
                    w.keyValueNot(KEY_TECH_LEVEL, mTechLevel, "");
                } else {
                    w.keyValue(KEY_TECH_LEVEL, "");
                }
            }
            if (mAttribute != SkillAttribute.IQ) {
                w.keyValue(KEY_ATTRIBUTE, Enums.toId(mAttribute));
            }
            w.keyValueNot(KEY_COLLEGE, mCollege, "");
            w.keyValueNot(KEY_POWER_SOURCE, mPowerSource, "");
            w.keyValueNot(KEY_SPELL_CLASS, mSpellClass, "");
            w.keyValueNot(KEY_RESIST, mResist, "");
            w.keyValueNot(KEY_CASTING_COST, mCastingCost, "");
            w.keyValueNot(KEY_MAINTENANCE_COST, mMaintenance, "");
            w.keyValueNot(KEY_CASTING_TIME, mCastingTime, "");
            w.keyValueNot(KEY_DURATION, mDuration, "");
            w.keyValueNot(KEY_POINTS, mPoints, 1);
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

    /** @return The level. */
    public int getLevel() {
        return mLevel.getLevel();
    }

    /** @return The relative level. */
    public int getRelativeLevel() {
        return mLevel.getRelativeLevel();
    }

    /** @return The calculated spell skill level. */
    private SkillLevel calculateLevelSelf() {
        return calculateLevel(getCharacter(), getPoints(), mAttribute, mDifficulty, mCollege, mPowerSource, mName, getCategories());
    }

    /**
     * Call to force an update of the level and relative level for this spell.
     *
     * @param notify Whether or not a notification should be issued on a change.
     */
    public void updateLevel(boolean notify) {
        SkillLevel savedLevel = mLevel;
        mLevel = calculateLevelSelf();
        if (notify && (savedLevel.isDifferentLevelThan(mLevel) || savedLevel.isDifferentRelativeLevelThan(mLevel))) {
            notifyOfChange();
        }
    }

    /**
     * Calculates the spell level.
     *
     * @param character   The character the spell will be attached to.
     * @param points      The number of points spent in the spell.
     * @param difficulty  The difficulty of the spell.
     * @param college     The college the spell belongs to.
     * @param powerSource The source of power for the spell.
     * @param name        The name of the spell.
     * @return The calculated spell level.
     */
    public static SkillLevel calculateLevel(GURPSCharacter character, int points, SkillAttribute attribute, SkillDifficulty difficulty, String college, String powerSource, String name, Set<String> categories) {
        StringBuilder toolTip       = new StringBuilder();
        int           relativeLevel = difficulty.getBaseRelativeLevel();
        int           level;

        if (character != null) {
            level = attribute.getBaseSkillLevel(character);
            if (difficulty == SkillDifficulty.W) {
                points /= 3;
            }
            if (points < 1) {
                level = -1;
                relativeLevel = 0;
            } else if (points == 1) {
                // mRelativeLevel is preset to this point value
            } else if (points < 4) {
                relativeLevel++;
            } else {
                relativeLevel += 1 + points / 4;
            }

            if (level != -1) {
                relativeLevel += getSpellBonusesFor(character, ID_COLLEGE, college, categories, toolTip);
                relativeLevel += getSpellBonusesFor(character, ID_POWER_SOURCE, powerSource, categories, toolTip);
                relativeLevel += getSpellBonusesFor(character, ID_NAME, name, categories, toolTip);
                level += relativeLevel;
            }
        } else {
            level = -1;
        }

        return new SkillLevel(level, relativeLevel, toolTip);
    }

    public static int getSpellBonusesFor(GURPSCharacter character, String id, String qualifier, Set<String> categories, StringBuilder toolTip) {
        int level = character.getIntegerBonusFor(id, toolTip);
        level += character.getIntegerBonusFor(id + '/' + qualifier.toLowerCase(), toolTip);
        level += character.getSpellComparedIntegerBonusFor(id + '*', qualifier, categories, toolTip);
        return level;
    }

    public static int getSpellPointBonusesFor(GURPSCharacter character, String id, String qualifier, Set<String> categories, StringBuilder toolTip) {
        int level = character.getIntegerBonusFor(id, toolTip);
        level += character.getIntegerBonusFor(id + '/' + qualifier.toLowerCase(), toolTip);
        level += character.getSpellPointComparedIntegerBonusFor(id + '*', qualifier, categories, toolTip);
        return level;
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

    /** @return The college. */
    public String getCollege() {
        return mCollege;
    }

    /**
     * @param college The college to set.
     * @return Whether it was changed.
     */
    public boolean setCollege(String college) {
        if (!mCollege.equals(college)) {
            mCollege = college;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The power source. */
    public String getPowerSource() {
        return mPowerSource;
    }

    /**
     * @param powerSource The college to set.
     * @return Whether it was changed.
     */
    public boolean setPowerSource(String powerSource) {
        if (!mPowerSource.equals(powerSource)) {
            mPowerSource = powerSource;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The class. */
    public String getSpellClass() {
        return mSpellClass;
    }

    /**
     * @param spellClass The class to set.
     * @return Whether it was modified.
     */
    public boolean setSpellClass(String spellClass) {
        if (!mSpellClass.equals(spellClass)) {
            mSpellClass = spellClass;
            return true;
        }
        return false;
    }

    /** @return The resistance. */
    public String getResist() {
        return mResist;
    }

    /**
     * @param resist The resistance to set.
     * @return Whether it was modified.
     */
    public boolean setResist(String resist) {
        if (!mResist.equals(resist)) {
            mResist = resist;
            return true;
        }
        return false;
    }

    /** @return The casting cost. */
    public String getCastingCost() {
        return mCastingCost;
    }

    /**
     * @param cost The casting cost to set.
     * @return Whether it was modified.
     */
    public boolean setCastingCost(String cost) {
        if (!mCastingCost.equals(cost)) {
            mCastingCost = cost;
            return true;
        }
        return false;
    }

    /** @return The maintainance cost. */
    public String getMaintenance() {
        return mMaintenance;
    }

    /**
     * @param cost The maintainance cost to set.
     * @return Whether it was modified.
     */
    public boolean setMaintenance(String cost) {
        if (!mMaintenance.equals(cost)) {
            mMaintenance = cost;
            return true;
        }
        return false;
    }

    /** @return The casting time. */
    public String getCastingTime() {
        return mCastingTime;
    }

    /**
     * @param castingTime The casting time to set.
     * @return Whether it was modified.
     */
    public boolean setCastingTime(String castingTime) {
        if (!mCastingTime.equals(castingTime)) {
            mCastingTime = castingTime;
            return true;
        }
        return false;
    }

    /** @return The duration. */
    public String getDuration() {
        return mDuration;
    }

    /**
     * @param duration The duration to set.
     * @return Whether it was modified.
     */
    public boolean setDuration(String duration) {
        if (!mDuration.equals(duration)) {
            mDuration = duration;
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
            StringBuilder tooltip    = new StringBuilder();
            Set<String>   categories = getCategories();
            getSpellPointBonusesFor(character, ID_POINTS_COLLEGE, getCollege(), categories, tooltip);
            getSpellPointBonusesFor(character, ID_POINTS_POWER_SOURCE, getPowerSource(), categories, tooltip);
            getSpellPointBonusesFor(character, ID_POINTS, getName(), categories, tooltip);
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
                if (row instanceof Spell) {
                    sum += ((Spell) row).getPoints();
                }
            }
            return sum;
        }
        int            points    = mPoints;
        GURPSCharacter character = getCharacter();
        if (character != null) {
            Set<String> categories = getCategories();
            points += getSpellPointBonusesFor(character, ID_POINTS_COLLEGE, getCollege(), categories, null);
            points += getSpellPointBonusesFor(character, ID_POINTS_POWER_SOURCE, getPowerSource(), categories, null);
            points += getSpellPointBonusesFor(character, ID_POINTS, getName(), categories, null);
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
     * @return Whether it was modified.
     */
    public boolean setRawPoints(int points) {
        if (mPoints != points) {
            mPoints = points;
            updateLevel(true);
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
            return true;
        }
        return false;
    }

    @Override
    public String getReferenceHighlight() {
        return getName();
    }

    @Override
    public Object getData(Column column) {
        return SpellColumn.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return SpellColumn.values()[column.getID()].getDataAsText(this);
    }

    /** @param text The combined attribute/difficulty to set. */
    // Copied from Skill class
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
        if (this instanceof RitualMagicSpell) {
            return (localized ? mDifficulty.toString() : mDifficulty.name());
        }
        return (localized ? mAttribute.toString() : mAttribute.name()) + "/" + (localized ? mDifficulty.toString() : mDifficulty.name());
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getName().toLowerCase().contains(text)) {
            return true;
        }
        if (getCollege().toLowerCase().contains(text)) {
            return true;
        }
        if (getSpellClass().toLowerCase().contains(text)) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    @Override
    public String toString() {
        StringBuilder builder = new StringBuilder();

        builder.append(getName());
        if (!canHaveChildren()) {
            String techLevel = getTechLevel();

            if (techLevel != null) {
                builder.append("/TL");
                if (!techLevel.isEmpty()) {
                    builder.append(techLevel);
                }
            }
        }
        return builder.toString();
    }

    @Override
    public RetinaIcon getIcon(boolean marker) {
        return marker ? Images.SPL_MARKER : Images.SPL_FILE;
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
     * @return Whether it was modified.
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

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new SpellEditor(this);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        extractNameables(set, mName);
        extractNameables(set, mCollege);
        extractNameables(set, mPowerSource);
        extractNameables(set, mCastingCost);
        extractNameables(set, mMaintenance);
        extractNameables(set, mCastingTime);
        extractNameables(set, mDuration);
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
        mCollege = nameNameables(map, mCollege);
        mPowerSource = nameNameables(map, mPowerSource);
        mSpellClass = nameNameables(map, mSpellClass);
        mCastingCost = nameNameables(map, mCastingCost);
        mMaintenance = nameNameables(map, mMaintenance);
        mCastingTime = nameNameables(map, mCastingTime);
        mDuration = nameNameables(map, mDuration);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.applyNameableKeys(map);
            }
        }
    }

    /** @return The default casting cost. */
    public static final String getDefaultCastingCost() {
        return "1";
    }

    /** @return The default casting time. */
    public static final String getDefaultCastingTime() {
        return I18n.Text("1 sec");
    }

    /** @return The default duration. */
    public static final String getDefaultDuration() {
        return I18n.Text("Instant");
    }

    /** @return The default power source. */
    public static final String getDefaultPowerSource() {
        return I18n.Text("Arcane");
    }

    /** @return The default spell class. */
    public static final String getDefaultSpellClass() {
        return I18n.Text("Regular");
    }

    @Override
    public String getToolTip(Column column) {
        return SpellColumn.values()[column.getID()].getToolTip(this);
    }

    public String getLevelToolTip() {
        return mLevel.getToolTip();
    }

    @Override
    public String getSecondaryText() {
        StringBuilder builder = new StringBuilder(super.getSecondaryText());
        String        rituals = getRituals();
        if (!rituals.isEmpty()) {
            if (!builder.isEmpty()) {
                builder.append("\n");
            }
            builder.append(rituals);
        }
        return builder.toString();
    }

    public String getRituals() {
        if (!((mDataFile instanceof GURPSCharacter) && ((GURPSCharacter) mDataFile).getSettings().showSpellAdj())) {
            return "";
        }
        int level = mLevel.getLevel();
        if (level < 10) {
            return I18n.Text("Ritual: need both hands and feet free and must speak; Time: 2x");
        }
        if (level < 15) {
            return I18n.Text("Ritual: speak quietly and make a gesture");
        }
        String ritual;
        String time = "";
        String cost = "";
        if (level < 20) {
            ritual = I18n.Text("speak a word or two OR make a small gesture"); // ; may move 1 yard per second while concentrating");
            if (!mSpellClass.toLowerCase().contains("blocking")) {
                cost = I18n.Text("; Cost: -1");
            }
        } else {
            ritual = I18n.Text("none");
            int adj = (level - 15) / 5;
            if (!mSpellClass.toLowerCase().contains("missile")) {
                time = String.format(I18n.Text("; Time: x1/%d, rounded up, min 1 sec"), Integer.valueOf(1 << adj));
            }
            if (!mSpellClass.toLowerCase().contains("blocking")) {
                cost = String.format(I18n.Text("; Cost: -%d"), Integer.valueOf(adj + 1));
            }
        }
        return I18n.Text("Ritual: ") + ritual + time + cost;
    }
}
