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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.Affects;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.ui.widget.outline.Switchable;
import com.trollworks.gcs.utility.Filtered;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;

/** A GURPS Advantage. */
public class Advantage extends ListRow implements HasSourceReference, Switchable {
    public static final  String KEY_ADVANTAGE           = "advantage";
    public static final  String KEY_ADVANTAGE_CONTAINER = "advantage_container";
    private static final String KEY_REFERENCE           = "reference";
    private static final String KEY_BASE_POINTS         = "base_points";
    private static final String KEY_POINTS_PER_LEVEL    = "points_per_level";
    private static final String KEY_LEVELS              = "levels";
    private static final String KEY_NAME                = "name";
    private static final String KEY_CR                  = "cr";
    private static final String KEY_USER_DESC           = "userdesc";
    private static final String KEY_DISABLED            = "disabled";
    private static final String KEY_ROUND_COST_DOWN     = "round_down";
    private static final String KEY_ALLOW_HALF_LEVELS   = "allow_half_levels";
    private static final String KEY_CONTAINER_TYPE      = "container_type";
    private static final String KEY_WEAPONS             = "weapons";
    private static final String KEY_MODIFIERS           = "modifiers";
    private static final String KEY_CR_ADJ              = "cr_adj";
    private static final String KEY_MENTAL              = "mental";
    private static final String KEY_PHYSICAL            = "physical";
    private static final String KEY_SOCIAL              = "social";
    private static final String KEY_EXOTIC              = "exotic";
    private static final String KEY_SUPERNATURAL        = "supernatural";
    private static final String TYPE_MENTAL             = "Mental";
    private static final String TYPE_PHYSICAL           = "Physical";
    private static final String TYPE_SOCIAL             = "Social";
    private static final String TYPE_EXOTIC             = "Exotic";
    private static final String TYPE_SUPERNATURAL       = "Supernatural";
    public static final  int    TYPE_MASK_MENTAL        = 1 << 0;
    public static final  int    TYPE_MASK_PHYSICAL      = 1 << 1;
    public static final  int    TYPE_MASK_SOCIAL        = 1 << 2;
    public static final  int    TYPE_MASK_EXOTIC        = 1 << 3;
    public static final  int    TYPE_MASK_SUPERNATURAL  = 1 << 4;

    private int                        mType;
    private String                     mName;
    private SelfControlRoll            mCR;
    private SelfControlRollAdjustments mCRAdj;
    private int                        mLevels;
    private int                        mPoints;
    private int                        mPointsPerLevel;
    private String                     mReference;
    private AdvantageContainerType     mContainerType;
    private List<WeaponStats>          mWeapons;
    private List<AdvantageModifier>    mModifiers;
    private String                     mUserDesc;
    private boolean                    mAllowHalfLevels;
    private boolean                    mHalfLevel;
    private boolean                    mRoundCostDown;
    private boolean                    mDisabled;

    /**
     * Creates a new advantage.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public Advantage(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
        mType = TYPE_MASK_PHYSICAL;
        mName = I18n.text("Advantage");
        mCR = SelfControlRoll.NONE_REQUIRED;
        mCRAdj = SelfControlRollAdjustments.NONE;
        mLevels = -1;
        mReference = "";
        mContainerType = AdvantageContainerType.GROUP;
        mWeapons = new ArrayList<>();
        mModifiers = new ArrayList<>();
        mUserDesc = "";
    }

    /**
     * Creates a clone of an existing advantage and associates it with the specified data file.
     *
     * @param dataFile  The data file to associate it with.
     * @param advantage The advantage to clone.
     * @param deep      Whether or not to clone the children, grandchildren, etc.
     */
    public Advantage(DataFile dataFile, Advantage advantage, boolean deep) {
        super(dataFile, advantage);
        mType = advantage.mType;
        mName = advantage.mName;
        mCR = advantage.mCR;
        mCRAdj = advantage.mCRAdj;
        mLevels = advantage.mLevels;
        mHalfLevel = advantage.mHalfLevel;
        mAllowHalfLevels = advantage.mAllowHalfLevels;
        mPoints = advantage.mPoints;
        mPointsPerLevel = advantage.mPointsPerLevel;
        mRoundCostDown = advantage.mRoundCostDown;
        mDisabled = advantage.mDisabled;
        mReference = advantage.mReference;
        mContainerType = advantage.mContainerType;
        mUserDesc = dataFile instanceof GURPSCharacter ? advantage.mUserDesc : "";
        mWeapons = new ArrayList<>(advantage.mWeapons.size());
        for (WeaponStats weapon : advantage.mWeapons) {
            if (weapon instanceof MeleeWeaponStats) {
                mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
            } else if (weapon instanceof RangedWeaponStats) {
                mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
            }
        }
        mModifiers = new ArrayList<>(advantage.mModifiers.size());
        for (AdvantageModifier modifier : advantage.mModifiers) {
            mModifiers.add(new AdvantageModifier(mDataFile, modifier, false));
        }
        if (deep) {
            int count = advantage.getChildCount();

            for (int i = 0; i < count; i++) {
                addChild(new Advantage(dataFile, (Advantage) advantage.getChild(i), true));
            }
        }
    }

    public Advantage(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, m.getString(DataFile.TYPE).equals(KEY_ADVANTAGE_CONTAINER));
        load(dataFile, m, state);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Advantage && super.isEquivalentTo(obj)) {
            Advantage row = (Advantage) obj;
            if (mType == row.mType && mLevels == row.mLevels && mHalfLevel == row.mHalfLevel && mPoints == row.mPoints && mPointsPerLevel == row.mPointsPerLevel && mDisabled == row.mDisabled && mRoundCostDown == row.mRoundCostDown && mAllowHalfLevels == row.mAllowHalfLevels && mContainerType == row.mContainerType && mCR == row.mCR && mCRAdj == row.mCRAdj && mName.equals(row.mName) && mReference.equals(row.mReference)) {
                if (mWeapons.equals(row.mWeapons)) {
                    return mModifiers.equals(row.mModifiers);
                }
            }
        }
        return false;
    }

    @Override
    public String getRowType() {
        return I18n.text("Advantage");
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_ADVANTAGE_CONTAINER : KEY_ADVANTAGE;
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mType = TYPE_MASK_PHYSICAL;
        mName = I18n.text("Advantage");
        mCR = SelfControlRoll.NONE_REQUIRED;
        mCRAdj = SelfControlRollAdjustments.NONE;
        mLevels = -1;
        mHalfLevel = false;
        mAllowHalfLevels = false;
        mReference = "";
        mContainerType = AdvantageContainerType.GROUP;
        mPoints = 0;
        mPointsPerLevel = 0;
        mRoundCostDown = false;
        mDisabled = false;
        mWeapons = new ArrayList<>();
        mModifiers = new ArrayList<>();
        mUserDesc = "";
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        mRoundCostDown = m.getBoolean(KEY_ROUND_COST_DOWN);
        mDisabled = m.getBoolean(KEY_DISABLED);
        mAllowHalfLevels = m.getBoolean(KEY_ALLOW_HALF_LEVELS);
        if (canHaveChildren()) {
            mContainerType = Enums.extract(m.getString(KEY_CONTAINER_TYPE), AdvantageContainerType.values(), AdvantageContainerType.GROUP);
        }
        mName = m.getString(KEY_NAME);
        mType = 0;
        if (!canHaveChildren()) {
            if (m.getBoolean(KEY_MENTAL)) {
                mType |= TYPE_MASK_MENTAL;
            }
            if (m.getBoolean(KEY_PHYSICAL)) {
                mType |= TYPE_MASK_PHYSICAL;
            }
            if (m.getBoolean(KEY_SOCIAL)) {
                mType |= TYPE_MASK_SOCIAL;
            }
            if (m.getBoolean(KEY_EXOTIC)) {
                mType |= TYPE_MASK_EXOTIC;
            }
            if (m.getBoolean(KEY_SUPERNATURAL)) {
                mType |= TYPE_MASK_SUPERNATURAL;
            }
            if (m.has(KEY_LEVELS)) {
                Fixed6 levels = new Fixed6(m.getString(KEY_LEVELS), false);
                mLevels = (int) levels.asLong();
                if (mAllowHalfLevels) {
                    mHalfLevel = levels.sub(new Fixed6(mLevels)).equals(new Fixed6(0.5));
                }
            }
            mPoints = m.getInt(KEY_BASE_POINTS);
            mPointsPerLevel = m.getInt(KEY_POINTS_PER_LEVEL);
            if (m.has(KEY_WEAPONS)) {
                WeaponStats.loadFromJSONArray(this, m.getArray(KEY_WEAPONS), mWeapons);
            }
        }
        if (m.has(KEY_CR)) {
            mCR = SelfControlRoll.getByCRValue(m.getInt(KEY_CR));
            if (m.has(KEY_CR_ADJ)) {
                mCRAdj = Enums.extract(m.getString(KEY_CR_ADJ), SelfControlRollAdjustments.values(), SelfControlRollAdjustments.NONE);
            }
        }
        if (m.has(KEY_MODIFIERS)) {
            JsonArray a     = m.getArray(KEY_MODIFIERS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mModifiers.add(new AdvantageModifier(getDataFile(), a.getMap(i), state));
            }
        }
        if (getDataFile() instanceof GURPSCharacter) {
            mUserDesc = m.getString(KEY_USER_DESC);
        }
        mReference = m.getString(KEY_REFERENCE);
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.TYPE);
            if (KEY_ADVANTAGE.equals(type) || KEY_ADVANTAGE_CONTAINER.equals(type)) {
                addChild(new Advantage(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        w.keyValueNot(KEY_ROUND_COST_DOWN, mRoundCostDown, false);
        w.keyValueNot(KEY_ALLOW_HALF_LEVELS, mAllowHalfLevels, false);
        w.keyValueNot(KEY_DISABLED, mDisabled, false);
        if (canHaveChildren() && mContainerType != AdvantageContainerType.GROUP) {
            w.keyValue(KEY_CONTAINER_TYPE, Enums.toId(mContainerType));
        }
        w.keyValue(KEY_NAME, mName);
        if (!canHaveChildren()) {
            w.keyValueNot(KEY_MENTAL, (mType & TYPE_MASK_MENTAL) != 0, false);
            w.keyValueNot(KEY_PHYSICAL, (mType & TYPE_MASK_PHYSICAL) != 0, false);
            w.keyValueNot(KEY_SOCIAL, (mType & TYPE_MASK_SOCIAL) != 0, false);
            w.keyValueNot(KEY_EXOTIC, (mType & TYPE_MASK_EXOTIC) != 0, false);
            w.keyValueNot(KEY_SUPERNATURAL, (mType & TYPE_MASK_SUPERNATURAL) != 0, false);
            if (mLevels != -1) {
                Fixed6 levels = new Fixed6(mLevels);
                if (mAllowHalfLevels && mHalfLevel) {
                    levels = levels.add(new Fixed6(0.5));
                }
                w.keyValue(KEY_LEVELS, levels.toString());
            }
            w.keyValueNot(KEY_BASE_POINTS, mPoints, 0);
            w.keyValueNot(KEY_POINTS_PER_LEVEL, mPointsPerLevel, 0);
            WeaponStats.saveList(w, KEY_WEAPONS, mWeapons);
        }
        if (mCR != SelfControlRoll.NONE_REQUIRED) {
            w.keyValue(KEY_CR, mCR.getCR());
            if (mCRAdj != SelfControlRollAdjustments.NONE) {
                w.keyValue(KEY_CR_ADJ, Enums.toId(mCRAdj));
            }
        }
        saveList(w, KEY_MODIFIERS, mModifiers, saveType);
        if (getDataFile() instanceof GURPSCharacter) {
            w.keyValueNot(KEY_USER_DESC, mUserDesc, "");
        }
        w.keyValueNot(KEY_REFERENCE, mReference, "");

        // Emit the calculated values for third parties
        w.key("calc");
        w.startMap();
        w.keyValue("points", getAdjustedPoints());
        w.endMap();
    }

    /** @return The container type. */
    public AdvantageContainerType getContainerType() {
        return mContainerType;
    }

    /**
     * @param type The container type to set.
     * @return Whether it was modified.
     */
    public boolean setContainerType(AdvantageContainerType type) {
        if (mContainerType != type) {
            mContainerType = type;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The type. */
    public int getType() {
        return mType;
    }

    /**
     * @param type The type to set.
     * @return Whether it was modified.
     */
    public boolean setType(int type) {
        if (mType != type) {
            mType = type;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.text("Advantage");
    }

    /** @return The name. */
    public String getName() {
        return mName;
    }

    public String getUserDesc() {
        return mUserDesc;
    }

    public boolean setUserDesc(String desc) {
        if (!mUserDesc.equals(desc)) {
            mUserDesc = desc;
            notifyOfChange();
            return true;
        }
        return false;
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

    /** @return The CR. */
    public SelfControlRoll getCR() {
        return mCR;
    }

    /**
     * @param cr The CR to set.
     * @return Whether it was changed.
     */
    public boolean setCR(SelfControlRoll cr) {
        if (mCR != cr) {
            mCR = cr;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The CR adjustment. */
    public SelfControlRollAdjustments getCRAdj() {
        return mCRAdj;
    }

    /**
     * @param crAdj The CR adjustment to set.
     * @return Whether it was changed.
     */
    public boolean setCRAdj(SelfControlRollAdjustments crAdj) {
        if (mCRAdj != crAdj) {
            mCRAdj = crAdj;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return Whether this advantage is leveled or not. */
    public boolean isLeveled() {
        return mLevels >= 0;
    }

    /** @return The levels. */
    public int getLevels() {
        return mLevels;
    }

    /**
     * @param levels The levels to set.
     * @return Whether it was modified.
     */
    public boolean setLevels(int levels) {
        if (mLevels != levels) {
            mLevels = levels;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return Whether there is at least half a level. */
    public boolean hasLevel() {
        return mLevels > 0 || mLevels == 0 && mHalfLevel && mAllowHalfLevels;
    }

    /** @return Whether there is a half level */
    public boolean hasHalfLevel() {
        return mHalfLevel;
    }

    /**
     * @param halfLevel The half level to set.
     * @return Whether it was modified.
     */
    public boolean setHalfLevel(boolean halfLevel) {
        if (mHalfLevel != halfLevel) {
            mHalfLevel = halfLevel;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @param factor The number of levels or half levels to set. */
    public void adjustLevel(int factor) {
        if (factor != 0) {
            if (mAllowHalfLevels) {
                int halfLevels = mLevels * 2 + (mHalfLevel ? 1 : 0) + factor;
                if (halfLevels < 0) {
                    halfLevels = 0;
                }
                setHalfLevel((halfLevels & 1) == 1);
                setLevels(halfLevels / 2);
            } else {
                setLevels(Math.max(mLevels + factor, 0));
            }
        }
    }

    /** @return The total points, taking levels into account. */
    public int getAdjustedPoints() {
        if (isDisabled()) {
            return 0;
        }
        if (canHaveChildren()) {
            int points = 0;
            if (mContainerType == AdvantageContainerType.ALTERNATIVE_ABILITIES) {
                List<Integer> values = new ArrayList<>();
                for (Advantage child : new FilteredIterator<>(getChildren(), Advantage.class)) {
                    int pts = child.getAdjustedPoints();
                    values.add(Integer.valueOf(pts));
                    if (pts > points) {
                        points = pts;
                    }
                }
                int     max   = points;
                boolean found = false;
                for (Integer one : values) {
                    int value = one.intValue();
                    if (!found && max == value) {
                        found = true;
                    } else {
                        points += applyRounding(calculateModifierPoints(value, 20), mRoundCostDown);
                    }
                }
            } else {
                for (Advantage child : new FilteredIterator<>(getChildren(), Advantage.class)) {
                    points += child.getAdjustedPoints();
                }
            }
            return points;
        }
        return getAdjustedPoints(mPoints, mLevels, mAllowHalfLevels && mHalfLevel, mPointsPerLevel, mCR, getAllModifiers(), mRoundCostDown);
    }

    private static int applyRounding(double value, boolean roundCostDown) {
        return (int) (roundCostDown ? Math.floor(value) : Math.ceil(value));
    }

    public boolean isDisabled() {
        return !isEnabled();
    }

    @Override
    public boolean isEnabled() {
        if (mDisabled) {
            return false;
        }
        Row parent = getParent();
        if (parent instanceof Switchable) {
            return ((Switchable) parent).isEnabled();
        }
        return true;
    }

    public boolean isSelfEnabled() {
        return !mDisabled;
    }

    public boolean setEnabled(boolean enabled) {
        if (mDisabled == enabled) {
            mDisabled = !enabled;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /**
     * @param basePoints     The base point cost.
     * @param levels         The number of levels.
     * @param halfLevel      Whether a half level is present.
     * @param pointsPerLevel The point cost per level.
     * @param cr             The {@link SelfControlRoll} to apply.
     * @param modifiers      The {@link AdvantageModifier}s to apply.
     * @param roundCostDown  Whether the point cost should be rounded down rather than up, as is
     *                       normal for most GURPS rules.
     * @return The total points, taking levels and modifiers into account.
     */
    public int getAdjustedPoints(int basePoints, int levels, boolean halfLevel, int pointsPerLevel, SelfControlRoll cr, Collection<AdvantageModifier> modifiers, boolean roundCostDown) {
        int    baseEnh    = 0;
        int    levelEnh   = 0;
        int    baseLim    = 0;
        int    levelLim   = 0;
        double multiplier = cr.getMultiplier();

        for (AdvantageModifier one : modifiers) {
            if (one.isEnabled()) {
                int modifier = one.getCostModifier();
                switch (one.getCostType()) {
                case PERCENTAGE:
                default:
                    switch (one.getAffects()) {
                    case TOTAL:
                    default:
                        if (modifier < 0) { // Limitation
                            baseLim += modifier;
                            levelLim += modifier;
                        } else { // Enhancement
                            baseEnh += modifier;
                            levelEnh += modifier;
                        }
                        break;
                    case BASE_ONLY:
                        if (modifier < 0) { // Limitation
                            baseLim += modifier;
                        } else { // Enhancement
                            baseEnh += modifier;
                        }
                        break;
                    case LEVELS_ONLY:
                        if (modifier < 0) { // Limitation
                            levelLim += modifier;
                        } else { // Enhancement
                            levelEnh += modifier;
                        }
                        break;
                    }
                    break;
                case POINTS:
                    if (one.getAffects() == Affects.LEVELS_ONLY) {
                        pointsPerLevel += modifier;
                    } else {
                        basePoints += modifier;
                    }
                    break;
                case MULTIPLIER:
                    multiplier *= one.getCostMultiplier();
                    break;
                }
            }
        }

        double modifiedBasePoints = basePoints;
        double leveledPoints      = pointsPerLevel * (levels + (halfLevel ? 0.5 : 0));
        if (baseEnh != 0 || baseLim != 0 || levelEnh != 0 || levelLim != 0) {
            if (getDataFile().getSheetSettings().useMultiplicativeModifiers()) {
                if (baseEnh == levelEnh && baseLim == levelLim) {
                    modifiedBasePoints = modifyPoints(modifiedBasePoints + leveledPoints, baseEnh);
                    modifiedBasePoints = modifyPoints(modifiedBasePoints, Math.max(baseLim, -80));
                } else {
                    modifiedBasePoints = modifyPoints(modifiedBasePoints, baseEnh);
                    modifiedBasePoints = modifyPoints(modifiedBasePoints, Math.max(baseLim, -80));
                    leveledPoints = modifyPoints(leveledPoints, levelEnh);
                    leveledPoints = modifyPoints(leveledPoints, Math.max(levelLim, -80));
                    modifiedBasePoints += leveledPoints;
                }
            } else {
                int baseMod  = Math.max(baseEnh + baseLim, -80);
                int levelMod = Math.max(levelEnh + levelLim, -80);
                modifiedBasePoints = baseMod == levelMod ? modifyPoints(modifiedBasePoints + leveledPoints, baseMod) : modifyPoints(modifiedBasePoints, baseMod) + modifyPoints(leveledPoints, levelMod);
            }
        } else {
            modifiedBasePoints += leveledPoints;
        }

        return applyRounding(modifiedBasePoints * multiplier, roundCostDown);
    }

    private static double modifyPoints(double points, int modifier) {
        return points + calculateModifierPoints(points, modifier);
    }

    private static double calculateModifierPoints(double points, int modifier) {
        return points * modifier / 100.0;
    }

    /** @return The points. */
    public int getPoints() {
        return mPoints;
    }

    /**
     * @param points The points to set.
     * @return Whether it was modified.
     */
    public boolean setPoints(int points) {
        if (mPoints != points) {
            mPoints = points;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The points per level. */
    public int getPointsPerLevel() {
        return mPointsPerLevel;
    }

    /**
     * @param points The points per level to set.
     * @return Whether it was modified.
     */
    public boolean setPointsPerLevel(int points) {
        if (mPointsPerLevel != points) {
            mPointsPerLevel = points;
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

    /**
     * @return Whether the point cost should be rounded down rather than up, as is normal for most
     *         GURPS rules.
     */
    public boolean shouldRoundCostDown() {
        return mRoundCostDown;
    }

    /**
     * @param shouldRoundDown Whether the point cost should be rounded down rather than up, as is
     *                        normal for most GURPS rules.
     * @return Whether it was modified.
     */
    public boolean setShouldRoundCostDown(boolean shouldRoundDown) {
        if (mRoundCostDown != shouldRoundDown) {
            mRoundCostDown = shouldRoundDown;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return Whether half levels are allowed */
    public boolean allowHalfLevels() {
        return mAllowHalfLevels;
    }

    public boolean setAllowHalfLevels(boolean allowHalfLevels) {
        if (mAllowHalfLevels != allowHalfLevels) {
            mAllowHalfLevels = allowHalfLevels;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public Object getData(Column column) {
        return AdvantageColumn.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return AdvantageColumn.values()[column.getID()].getDataAsText(this);
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getName().toLowerCase().contains(text)) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    /** @return The type as a text string. */
    public String getTypeAsText() {
        if (!canHaveChildren()) {
            String        separator = ", ";
            StringBuilder buffer    = new StringBuilder();
            int           type      = getType();
            if ((type & TYPE_MASK_MENTAL) != 0) {
                buffer.append(TYPE_MENTAL);
            }
            if ((type & TYPE_MASK_PHYSICAL) != 0) {
                if (!buffer.isEmpty()) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_PHYSICAL);
            }
            if ((type & TYPE_MASK_SOCIAL) != 0) {
                if (!buffer.isEmpty()) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_SOCIAL);
            }
            if ((type & TYPE_MASK_EXOTIC) != 0) {
                if (!buffer.isEmpty()) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_EXOTIC);
            }
            if ((type & TYPE_MASK_SUPERNATURAL) != 0) {
                if (!buffer.isEmpty()) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_SUPERNATURAL);
            }
            return buffer.toString();
        }
        return "";
    }

    @Override
    public String toString() {
        StringBuilder builder = new StringBuilder();
        builder.append(getName());
        if (!canHaveChildren()) {
            boolean halfLevel = mAllowHalfLevels && mHalfLevel;
            if (mLevels > 0 || halfLevel) {
                builder.append(' ');
                if (mLevels > 0) {
                    builder.append(mLevels);
                }
                if (halfLevel) {
                    builder.append('½');
                }
            }
        }
        return builder.toString();
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

    @Override
    public RetinaIcon getIcon(boolean marker) {
        return marker ? Images.ADQ_MARKER : Images.ADQ_FILE;
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new AdvantageEditor(this);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        extractNameables(set, mName);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.fillWithNameableKeys(set);
            }
        }
        for (AdvantageModifier modifier : mModifiers) {
            modifier.fillWithNameableKeys(set);
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        super.applyNameableKeys(map);
        mName = nameNameables(map, mName);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.applyNameableKeys(map);
            }
        }
        for (AdvantageModifier modifier : mModifiers) {
            modifier.applyNameableKeys(map);
        }
    }

    /** @return The modifiers. */
    public List<AdvantageModifier> getModifiers() {
        return Collections.unmodifiableList(mModifiers);
    }

    /** @return The modifiers including those inherited from parent row. */
    public List<AdvantageModifier> getAllModifiers() {
        List<AdvantageModifier> allModifiers = new ArrayList<>(mModifiers);
        if (getParent() != null) {
            allModifiers.addAll(((Advantage) getParent()).getAllModifiers());
        }
        return Collections.unmodifiableList(allModifiers);
    }

    /** @param modifiers The value to set for modifiers. */
    public void setModifiers(List<? extends Modifier> modifiers) {
        List<AdvantageModifier> in = Filtered.list(modifiers, AdvantageModifier.class);
        if (!mModifiers.equals(in)) {
            mModifiers = in;
            notifyOfChange();
            update();
        }
    }

    /**
     * @param name The name to match against. Case-insensitive.
     * @return The first modifier that matches the name.
     */
    public AdvantageModifier getActiveModifierFor(String name) {
        for (AdvantageModifier m : getModifiers()) {
            if (m.isEnabled() && m.getName().equalsIgnoreCase(name)) {
                return m;
            }
        }
        return null;
    }

    private static final String MODIFIER_SEPARATOR = "; ";

    @Override
    public String getModifierNotes() {
        StringBuilder builder = new StringBuilder();
        if (mCR != SelfControlRoll.NONE_REQUIRED) {
            builder.append(mCR);
            if (mCRAdj != SelfControlRollAdjustments.NONE) {
                builder.append(", ");
                builder.append(mCRAdj.getDescription(getCR()));
            }
            builder.append(MODIFIER_SEPARATOR);
        }
        for (AdvantageModifier modifier : mModifiers) {
            if (modifier.isEnabled()) {
                builder.append(modifier.getFullDescription());
                builder.append(MODIFIER_SEPARATOR);
            }
        }
        if (!builder.isEmpty()) {
            // Remove the trailing MODIFIER_SEPARATOR
            builder.setLength(builder.length() - MODIFIER_SEPARATOR.length());
        }
        return builder.toString();
    }

    /** @return The "secondary" text, the text display below an Advantage. */
    @Override
    protected String getSecondaryText() {
        StringBuilder builder = new StringBuilder();
        if (getDataFile().getSheetSettings().userDescriptionDisplay().inline()) {
            String txt = getUserDesc();
            if (!txt.isBlank()) {
                builder.append(txt);
            }
        }
        String txt = super.getSecondaryText();
        if (!txt.isBlank()) {
            if (!builder.isEmpty()) {
                builder.append("\n");
            }
            builder.append(super.getSecondaryText());
        }
        return builder.toString();
    }

    @Override
    public String getToolTip(Column column) {
        return AdvantageColumn.values()[column.getID()].getToolTip(this);
    }
}
