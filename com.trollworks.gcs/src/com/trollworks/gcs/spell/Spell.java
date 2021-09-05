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

import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDifficulty;
import com.trollworks.gcs.skill.SkillLevel;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.ID;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Numbers;
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
import java.util.regex.Pattern;
import javax.swing.Icon;

/** A GURPS Spell. */
public class Spell extends ListRow implements HasSourceReference {
    private static final   int    COLLEGE_LIST_VERSION = 2; // First version with college lists (post v4.29.1)
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
    private static final   String KEY_DIFFICULTY       = "difficulty";
    private static final   String KEY_WEAPONS          = "weapons";

    public static final String ID_NAME                = "spell.name";
    public static final String ID_COLLEGE             = "spell.college";
    public static final String ID_POWER_SOURCE        = "spell.power_source";
    public static final String ID_POINTS              = "spell.points";
    public static final String ID_POINTS_COLLEGE      = "spell.college.points";
    public static final String ID_POINTS_POWER_SOURCE = "spell.power_source.points";

    private static final Pattern COLLEGE_OR        = Pattern.compile("(\\s+or\\s+)|/", Pattern.CASE_INSENSITIVE);
    private static final Pattern LINE_FEED_PATTERN = Pattern.compile("\n");

    private   String            mName;
    private   String            mTechLevel;
    private   List<String>      mColleges;
    private   String            mPowerSource;
    private   String            mSpellClass;
    private   String            mResist;
    private   String            mCastingCost;
    private   String            mMaintenance;
    private   String            mCastingTime;
    private   String            mDuration;
    protected int               mPoints;
    protected SkillLevel        mLevel;
    private   String            mAttribute;
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
        mName = I18n.text("Spell");
        mAttribute = Skill.getDefaultAttribute("iq");
        mDifficulty = SkillDifficulty.H;
        mTechLevel = null;
        mColleges = new ArrayList<>();
        mPowerSource = isContainer ? "" : getDefaultPowerSource();
        mSpellClass = isContainer ? "" : getDefaultSpellClass();
        mResist = "";
        mCastingCost = isContainer ? "" : getDefaultCastingCost();
        mMaintenance = "";
        mCastingTime = isContainer ? "" : getDefaultCastingTime();
        mDuration = isContainer ? "" : getDefaultDuration();
        mPoints = 1;
        mReference = "";
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
        mDifficulty = spell.mDifficulty;
        mTechLevel = spell.mTechLevel;
        mColleges = new ArrayList<>(spell.mColleges);
        mPowerSource = spell.mPowerSource;
        mSpellClass = spell.mSpellClass;
        mResist = spell.mResist;
        mCastingCost = spell.mCastingCost;
        mMaintenance = spell.mMaintenance;
        mCastingTime = spell.mCastingTime;
        mDuration = spell.mDuration;
        mPoints = forSheet ? spell.mPoints : 1;
        mReference = spell.mReference;
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
        this(dataFile, m.getString(DataFile.TYPE).equals(KEY_SPELL_CONTAINER));
        load(dataFile, m, state);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Spell && super.isEquivalentTo(obj)) {
            Spell row = (Spell) obj;
            if (mDifficulty != row.mDifficulty) {
                return false;
            }
            if (mPoints != row.mPoints) {
                return false;
            }
            if (!mAttribute.equals(row.mAttribute)) {
                return false;
            }
            if (!mLevel.isSameLevelAs(row.mLevel)) {
                return false;
            }
            if (!Objects.equals(mTechLevel, row.mTechLevel)) {
                return false;
            }
            if (!mName.equals(row.mName)) {
                return false;
            }
            if (!mColleges.equals(row.mColleges)) {
                return false;
            }
            if (!mPowerSource.equals(row.mPowerSource)) {
                return false;
            }
            if (!mSpellClass.equals(row.mSpellClass)) {
                return false;
            }
            if (!mResist.equals(row.mResist)) {
                return false;
            }
            if (!mReference.equals(row.mReference)) {
                return false;
            }
            if (!mCastingCost.equals(row.mCastingCost)) {
                return false;
            }
            if (!mMaintenance.equals(row.mMaintenance)) {
                return false;
            }
            if (!mCastingTime.equals(row.mCastingTime)) {
                return false;
            }
            if (!mDuration.equals(row.mDuration)) {
                return false;
            }
            return mWeapons.equals(row.mWeapons);
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.text("Spell");
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_SPELL_CONTAINER : KEY_SPELL;
    }

    @Override
    public String getRowType() {
        return I18n.text("Spell");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        boolean isContainer = canHaveChildren();
        super.prepareForLoad(state);
        mName = I18n.text("Spell");
        mAttribute = Skill.getDefaultAttribute("iq");
        mTechLevel = null;
        mColleges = new ArrayList<>();
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
            if (state.mDataFileVersion >= COLLEGE_LIST_VERSION) {
                JsonArray a    = m.getArray(KEY_COLLEGE);
                int       size = a.size();
                for (int i = 0; i < size; i++) {
                    String s = a.getString(i);
                    if (!s.isBlank()) {
                        mColleges.add(s);
                    }
                }
                Collections.sort(mColleges);
            } else {
                // Legacy (v4.29.1 and earlier)
                String s = m.getString(KEY_COLLEGE);
                if (!s.isBlank()) {
                    for (String college : COLLEGE_OR.split(s)) {
                        if (!college.isBlank()) {
                            mColleges.add(college);
                        }
                    }
                }
            }
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
            String type = m.getString(DataFile.TYPE);
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
            if (!mColleges.isEmpty()) {
                w.key(KEY_COLLEGE);
                w.startArray();
                for (String college : mColleges) {
                    w.value(college);
                }
                w.endArray();
            }
            w.keyValueNot(KEY_POWER_SOURCE, mPowerSource, "");
            w.keyValueNot(KEY_SPELL_CLASS, mSpellClass, "");
            w.keyValueNot(KEY_RESIST, mResist, "");
            w.keyValueNot(KEY_CASTING_COST, mCastingCost, "");
            w.keyValueNot(KEY_MAINTENANCE_COST, mMaintenance, "");
            w.keyValueNot(KEY_CASTING_TIME, mCastingTime, "");
            w.keyValueNot(KEY_DURATION, mDuration, "");
            w.keyValue(KEY_POINTS, mPoints);
            WeaponStats.saveList(w, KEY_WEAPONS, mWeapons);

            // Emit the calculated values for third parties
            int level = getLevel();
            if (level > 0) {
                w.key("calc");
                w.startMap();
                w.keyValue("level", level);
                StringBuilder builder = new StringBuilder();
                int           rsl     = getAdjustedRelativeLevel();
                if (rsl == Integer.MIN_VALUE) {
                    builder.append("-");
                } else {
                    if (!(this instanceof RitualMagicSpell)) {
                        builder.append(Skill.resolveAttributeName(getDataFile(), getAttribute()));
                    }
                    builder.append(Numbers.formatWithForcedSign(rsl));
                }
                w.keyValue("rsl", builder.toString());
                w.endMap();
            }
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
        return calculateLevel(getCharacter(), getPoints(), mAttribute, mDifficulty, mColleges, mPowerSource, mName, getCategories());
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
     * @param colleges    The colleges the spell belongs to.
     * @param powerSource The source of power for the spell.
     * @param name        The name of the spell.
     * @return The calculated spell level.
     */
    public static SkillLevel calculateLevel(GURPSCharacter character, int points, String attribute, SkillDifficulty difficulty, List<String> colleges, String powerSource, String name, Set<String> categories) {
        StringBuilder toolTip       = new StringBuilder();
        int           relativeLevel = difficulty.getBaseRelativeLevel();
        int           level;

        if (character != null) {
            level = Skill.resolveAttribute(character, attribute);
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
                relativeLevel += getBestCollegeSpellBonus(character, categories, colleges, toolTip);
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

    public static int getBestCollegeSpellBonus(GURPSCharacter character, Set<String> categories, List<String> colleges, StringBuilder tooltip) {
        int    best        = Integer.MIN_VALUE;
        String bestTooltip = "";
        for (String college : colleges) {
            StringBuilder buffer = tooltip != null ? new StringBuilder() : null;
            int           pts    = getSpellBonusesFor(character, ID_COLLEGE, college, categories, buffer);
            if (best < pts) {
                best = pts;
                if (buffer != null) {
                    bestTooltip = buffer.toString();
                }
            }
        }
        if (tooltip != null) {
            tooltip.append(bestTooltip);
        }
        return best == Integer.MIN_VALUE ? 0 : best;
    }

    public int getBestCollegeSpellPointBonus(StringBuilder tooltip) {
        GURPSCharacter character   = getCharacter();
        Set<String>    categories  = getCategories();
        int            best        = Integer.MIN_VALUE;
        String         bestTooltip = "";
        for (String college : getColleges()) {
            StringBuilder buffer = tooltip != null ? new StringBuilder() : null;
            int           pts    = getSpellPointBonusesFor(character, ID_POINTS_COLLEGE, college, categories, buffer);
            if (best < pts) {
                best = pts;
                if (buffer != null) {
                    bestTooltip = buffer.toString();
                }
            }
        }
        if (tooltip != null) {
            tooltip.append(bestTooltip);
        }
        return best == Integer.MIN_VALUE ? 0 : best;
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

    /** @return The colleges. */
    public List<String> getColleges() {
        return mColleges;
    }

    /**
     * @param colleges The colleges to set.
     * @return Whether it was changed.
     */
    public boolean setColleges(List<String> colleges) {
        colleges = new ArrayList<>(colleges);
        Collections.sort(colleges);
        if (!mColleges.equals(colleges)) {
            mColleges = colleges;
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
            return I18n.text("The sum of the points spent by children of this container");
        }
        GURPSCharacter character = getCharacter();
        if (character != null) {
            StringBuilder tooltip    = new StringBuilder();
            Set<String>   categories = getCategories();
            getBestCollegeSpellPointBonus(tooltip);
            getSpellPointBonusesFor(character, ID_POINTS_POWER_SOURCE, getPowerSource(), categories, tooltip);
            getSpellPointBonusesFor(character, ID_POINTS, getName(), categories, tooltip);
            if (!tooltip.isEmpty()) {
                return I18n.text("Includes modifiers from") + tooltip;
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
            points += getBestCollegeSpellPointBonus(null);
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
    // Copied from Skill class (mostly)
    public void setDifficultyFromText(String text) {
        String[]        parts      = text.split("/", 2);
        SkillDifficulty difficulty = SkillDifficulty.A;
        if (parts.length == 2) {
            String diffText = parts[1].trim();
            for (SkillDifficulty d : SkillDifficulty.values()) {
                if (d.name().equalsIgnoreCase(diffText)) {
                    difficulty = d;
                    break;
                }
            }
        }
        String attrText;
        if (parts.length > 0) {
            attrText = parts[0].trim();
        } else {
            attrText = Skill.getDefaultAttribute("iq");
        }
        AttributeDef attr = null;
        for (AttributeDef attrDef : AttributeDef.getOrdered(getDataFile().getSheetSettings().getAttributes())) {
            if (attrDef.getID().equalsIgnoreCase(attrText)) {
                attr = attrDef;
                break;
            }
        }
        if (attr == null) {
            for (AttributeDef attrDef : AttributeDef.getOrdered(getDataFile().getSheetSettings().getAttributes())) {
                if (attrDef.getName().equalsIgnoreCase(attrText)) {
                    attr = attrDef;
                    break;
                }
            }
        }
        if (attr != null) {
            attrText = attr.getID();
        }
        setDifficulty(attrText, difficulty);
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
            return (localized ? mDifficulty.toString() : mDifficulty.name().toLowerCase());
        }
        if (localized) {
            return Skill.resolveAttributeName(getDataFile(), mAttribute) + "/" + mDifficulty.toString();
        }
        return mAttribute + "/" + mDifficulty.name().toLowerCase();
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getName().toLowerCase().contains(text)) {
            return true;
        }
        if (getSpellClass().toLowerCase().contains(text)) {
            return true;
        }
        for (String college : getColleges()) {
            if (college.toLowerCase().contains(text)) {
                return true;
            }
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
    public Icon getIcon() {
        return FileType.SPELL.getIcon();
    }

    /** @return The attribute. */
    public String getAttribute() {
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
    public boolean setDifficulty(String attribute, SkillDifficulty difficulty) {
        if (mDifficulty != difficulty || !mAttribute.equals(attribute)) {
            mAttribute = ID.sanitize(attribute, null, true);
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
        for (String college : getColleges()) {
            extractNameables(set, college);
        }
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
        List<String> colleges = new ArrayList<>();
        for (String college : getColleges()) {
            colleges.add(nameNameables(map, college));
        }
        setColleges(colleges);
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
        return I18n.text("1 sec");
    }

    /** @return The default duration. */
    public static final String getDefaultDuration() {
        return I18n.text("Instant");
    }

    /** @return The default power source. */
    public static final String getDefaultPowerSource() {
        return I18n.text("Arcane");
    }

    /** @return The default spell class. */
    public static final String getDefaultSpellClass() {
        return I18n.text("Regular");
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
        if (getDataFile().getSheetSettings().skillLevelAdjustmentsDisplay().inline()) {
            String levelTooltip = getLevelToolTip();
            if (levelTooltip != null && !SkillLevel.getNoAdditionalModifiers().equals(levelTooltip)) {
                if (!builder.isEmpty()) {
                    builder.append("\n");
                }
                levelTooltip = LINE_FEED_PATTERN.matcher(levelTooltip).replaceAll(", ");
                String includesPrefix = SkillLevel.getIncludesModifiersFrom();
                if (levelTooltip.startsWith(includesPrefix + ",")) {
                    levelTooltip = includesPrefix + ":" + levelTooltip.substring(includesPrefix.length() + 1);
                }
                builder.append(levelTooltip);
            }
        }
        return builder.toString();
    }

    public String getRituals() {
        if (!((mDataFile instanceof GURPSCharacter) && mDataFile.getSheetSettings().showSpellAdj())) {
            return "";
        }
        int level = mLevel.getLevel();
        if (level < 10) {
            return I18n.text("Ritual: need both hands and feet free and must speak; Time: 2x");
        }
        if (level < 15) {
            return I18n.text("Ritual: speak quietly and make a gesture");
        }
        String ritual;
        String time = "";
        String cost = "";
        if (level < 20) {
            ritual = I18n.text("speak a word or two OR make a small gesture"); // ; may move 1 yard per second while concentrating");
            if (!mSpellClass.toLowerCase().contains("blocking")) {
                cost = I18n.text("; Cost: -1");
            }
        } else {
            ritual = I18n.text("none");
            int adj = (level - 15) / 5;
            if (!mSpellClass.toLowerCase().contains("missile")) {
                time = String.format(I18n.text("; Time: x1/%d, rounded up, min 1 sec"), Integer.valueOf(1 << adj));
            }
            if (!mSpellClass.toLowerCase().contains("blocking")) {
                cost = String.format(I18n.text("; Cost: -%d"), Integer.valueOf(adj + 1));
            }
        }
        return I18n.text("Ritual: ") + ritual + time + cost;
    }

    public int getAdjustedRelativeLevel() {
        if (!canHaveChildren()) {
            if (getCharacter() != null) {
                if (getLevel() < 0) {
                    return Integer.MIN_VALUE;
                }
                return getRelativeLevel();
            }
        } else if (getTemplate() != null) {
            int points = getPoints();
            if (points > 0) {
                SkillDifficulty difficulty = getDifficulty();
                int             level      = difficulty.getBaseRelativeLevel();
                if (difficulty == SkillDifficulty.W) {
                    points /= 3;
                }
                if (points > 1) {
                    if (points < 4) {
                        level++;
                    } else {
                        level += 1 + points / 4;
                    }
                }
                return level;
            }
        }
        return Integer.MIN_VALUE;
    }
}
