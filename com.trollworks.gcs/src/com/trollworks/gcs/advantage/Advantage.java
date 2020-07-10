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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.modifier.AdvantageModifier;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.ui.widget.outline.Switchable;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.xml.XMLReader;
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
    private static final int                        CURRENT_JSON_VERSION       = 1;
    private static final int                        CURRENT_VERSION            = 4;
    /** The XML tag used for items. */
    public static final  String                     TAG_ADVANTAGE              = "advantage";
    /** The XML tag used for containers. */
    public static final  String                     TAG_ADVANTAGE_CONTAINER    = "advantage_container";
    private static final String                     TAG_REFERENCE              = "reference";
    private static final String                     TAG_BASE_POINTS            = "base_points";
    private static final String                     TAG_POINTS_PER_LEVEL       = "points_per_level";
    private static final String                     TAG_LEVELS                 = "levels";
    private static final String                     TAG_TYPE                   = "type";
    private static final String                     TAG_NAME                   = "name";
    private static final String                     TAG_CR                     = "cr";
    private static final String                     TAG_USER_DESC              = "userdesc";
    private static final String                     TYPE_MENTAL                = "Mental";
    private static final String                     TYPE_PHYSICAL              = "Physical";
    private static final String                     TYPE_SOCIAL                = "Social";
    private static final String                     TYPE_EXOTIC                = "Exotic";
    private static final String                     TYPE_SUPERNATURAL          = "Supernatural";
    private static final String                     ATTR_DISABLED              = "disabled";
    private static final String                     ATTR_ROUND_COST_DOWN       = "round_down";
    private static final String                     ATTR_ALLOW_HALF_LEVELS     = "allow_half_levels";
    private static final String                     ATTR_HALF_LEVEL            = "half_level";
    private static final String                     KEY_CONTAINER_TYPE         = "container_type";
    private static final String                     KEY_WEAPONS                = "weapons";
    private static final String                     KEY_MODIFIERS              = "modifiers";
    private static final String                     KEY_CR_ADJ                 = "cr_adj";
    private static final String                     KEY_MENTAL                 = "mental";
    private static final String                     KEY_PHYSICAL               = "physical";
    private static final String                     KEY_SOCIAL                 = "social";
    private static final String                     KEY_EXOTIC                 = "exotic";
    private static final String                     KEY_SUPERNATURAL           = "supernatural";
    /** The prefix used in front of all IDs for the advantages. */
    public static final  String                     PREFIX                     = GURPSCharacter.CHARACTER_PREFIX + "advantage" + Notifier.SEPARATOR;
    /** The field ID for type changes. */
    public static final  String                     ID_TYPE                    = PREFIX + "Type";
    /** The field ID for container type changes. */
    public static final  String                     ID_CONTAINER_TYPE          = PREFIX + "ContainerType";
    /** The field ID for name changes. */
    public static final  String                     ID_NAME                    = PREFIX + "Name";
    /** The field ID for CR changes. */
    public static final  String                     ID_CR                      = PREFIX + "CR";
    /** The field ID for level changes. */
    public static final  String                     ID_LEVELS                  = PREFIX + "Levels";
    /** The field ID for half level. */
    public static final  String                     ID_HALF_LEVEL              = PREFIX + "HalfLevel";
    /** The field ID for round cost down changes. */
    public static final  String                     ID_ROUND_COST_DOWN         = PREFIX + "RoundCostDown";
    /** The field ID for disabled changes. */
    public static final  String                     ID_DISABLED                = PREFIX + "Disabled";
    /** The field ID for allowing half levels. */
    public static final  String                     ID_ALLOW_HALF_LEVELS       = PREFIX + "AllowHalfLevels";
    /** The field ID for point changes. */
    public static final  String                     ID_POINTS                  = PREFIX + "Points";
    /** The field ID for page reference changes. */
    public static final  String                     ID_REFERENCE               = PREFIX + "Reference";
    /** The field ID for when the categories change. */
    public static final  String                     ID_CATEGORY                = PREFIX + "Category";
    /** The field ID for when the row hierarchy changes. */
    public static final  String                     ID_LIST_CHANGED            = PREFIX + "ListChanged";
    /** The field ID for when the advantage becomes or stops being a weapon. */
    public static final  String                     ID_WEAPON_STATUS_CHANGED   = PREFIX + "WeaponStatus";
    /** The field ID for when the advantage gets Modifiers. */
    public static final  String                     ID_MODIFIER_STATUS_CHANGED = PREFIX + "Modifier";
    /** The field ID for user description changes. */
    public static final  String                     ID_USER_DESC               = PREFIX + "UserDesc";
    /** The type mask for mental advantages. */
    public static final  int                        TYPE_MASK_MENTAL           = 1 << 0;
    /** The type mask for physical advantages. */
    public static final  int                        TYPE_MASK_PHYSICAL         = 1 << 1;
    /** The type mask for social advantages. */
    public static final  int                        TYPE_MASK_SOCIAL           = 1 << 2;
    /** The type mask for exotic advantages. */
    public static final  int                        TYPE_MASK_EXOTIC           = 1 << 3;
    /** The type mask for supernatural advantages. */
    public static final  int                        TYPE_MASK_SUPERNATURAL     = 1 << 4;
    private              int                        mType;
    private              String                     mName;
    private              SelfControlRoll            mCR;
    private              SelfControlRollAdjustments mCRAdj;
    private              int                        mLevels;
    private              boolean                    mAllowHalfLevels;
    private              boolean                    mHalfLevel;
    private              int                        mPoints;
    private              int                        mPointsPerLevel;
    private              String                     mReference;
    private              AdvantageContainerType     mContainerType;
    private              List<WeaponStats>          mWeapons;
    private              List<AdvantageModifier>    mModifiers;
    private              boolean                    mRoundCostDown;
    private              boolean                    mDisabled;
    private              String                     mUserDesc;

    /**
     * Creates a new advantage.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public Advantage(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
        mType = TYPE_MASK_PHYSICAL;
        mName = I18n.Text("Advantage");
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
        this(dataFile, m.getString(DataFile.KEY_TYPE).equals(TAG_ADVANTAGE_CONTAINER));
        load(m, state);
    }

    /**
     * Loads an advantage and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param reader   The XML reader to load from.
     * @param state    The {@link LoadState} to use.
     */
    public Advantage(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
        this(dataFile, TAG_ADVANTAGE_CONTAINER.equals(reader.getName()));
        load(reader, state);
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
    public String getListChangedID() {
        return ID_LIST_CHANGED;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Advantage");
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? TAG_ADVANTAGE_CONTAINER : TAG_ADVANTAGE;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public String getXMLTagName() {
        return canHaveChildren() ? TAG_ADVANTAGE_CONTAINER : TAG_ADVANTAGE;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mType = TYPE_MASK_PHYSICAL;
        mName = I18n.Text("Advantage");
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
    protected void loadAttributes(XMLReader reader, LoadState state) {
        super.loadAttributes(reader, state);
        mRoundCostDown = reader.isAttributeSet(ATTR_ROUND_COST_DOWN);
        mDisabled = reader.isAttributeSet(ATTR_DISABLED);
        mAllowHalfLevels = reader.isAttributeSet(ATTR_ALLOW_HALF_LEVELS);
        if (canHaveChildren()) {
            mContainerType = Enums.extract(reader.getAttribute(TAG_TYPE), AdvantageContainerType.values(), AdvantageContainerType.GROUP);
        }
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (TAG_NAME.equals(name)) {
            mName = reader.readText().replace("\n", " ");
        } else if (TAG_CR.equals(name)) {
            mCRAdj = Enums.extract(reader.getAttribute(SelfControlRoll.ATTR_ADJUSTMENT), SelfControlRollAdjustments.values(), SelfControlRollAdjustments.NONE);
            mCR = SelfControlRoll.get(reader.readText());
        } else if (TAG_REFERENCE.equals(name)) {
            mReference = reader.readText().replace("\n", " ");
        } else if (!state.mForUndo && (TAG_ADVANTAGE.equals(name) || TAG_ADVANTAGE_CONTAINER.equals(name))) {
            addChild(new Advantage(mDataFile, reader, state));
        } else if (AdvantageModifier.TAG_MODIFIER.equals(name)) {
            mModifiers.add(new AdvantageModifier(getDataFile(), reader, state));
        } else if (TAG_USER_DESC.equals(name)) {
            if (getDataFile() instanceof GURPSCharacter) {
                mUserDesc = Text.standardizeLineEndings(reader.readText());
            }
        } else if (!canHaveChildren()) {
            if (TAG_TYPE.equals(name)) {
                mType = getTypeFromText(reader.readText());
            } else if (TAG_LEVELS.equals(name)) {
                // Read the attribute first as next operation clears attribute map
                mHalfLevel = mAllowHalfLevels && reader.isAttributeSet(ATTR_HALF_LEVEL);
                mLevels = reader.readInteger(-1);
            } else if (TAG_BASE_POINTS.equals(name)) {
                mPoints = reader.readInteger(0);
            } else if (TAG_POINTS_PER_LEVEL.equals(name)) {
                mPointsPerLevel = reader.readInteger(0);
            } else if (MeleeWeaponStats.TAG_ROOT.equals(name)) {
                mWeapons.add(new MeleeWeaponStats(this, reader));
            } else if (RangedWeaponStats.TAG_ROOT.equals(name)) {
                mWeapons.add(new RangedWeaponStats(this, reader));
            } else {
                super.loadSubElement(reader, state);
            }
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        mRoundCostDown = m.getBoolean(ATTR_ROUND_COST_DOWN);
        mDisabled = m.getBoolean(ATTR_DISABLED);
        mAllowHalfLevels = m.getBoolean(ATTR_ALLOW_HALF_LEVELS);
        if (canHaveChildren()) {
            mContainerType = Enums.extract(m.getString(TAG_TYPE), AdvantageContainerType.values(), AdvantageContainerType.GROUP);
        }
        mName = m.getString(TAG_NAME);
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
            if (m.has(TAG_LEVELS)) {
                Fixed6 levels = new Fixed6(m.getString(TAG_LEVELS), false);
                mLevels = (int) levels.asLong();
                if (mAllowHalfLevels) {
                    mHalfLevel = levels.sub(new Fixed6(mLevels)).equals(new Fixed6(0.5));
                }
            }
            mPoints = m.getInt(TAG_BASE_POINTS);
            mPointsPerLevel = m.getInt(TAG_POINTS_PER_LEVEL);
            if (m.has(KEY_WEAPONS)) {
                WeaponStats.loadFromJSONArray(this, m.getArray(KEY_WEAPONS), mWeapons);
            }
        }
        if (m.has(TAG_CR)) {
            mCR = SelfControlRoll.getByCRValue(m.getInt(TAG_CR));
            if (m.has(KEY_CR_ADJ)) {
                mCRAdj = Enums.extract(m.getString(SelfControlRoll.ATTR_ADJUSTMENT), SelfControlRollAdjustments.values(), SelfControlRollAdjustments.NONE);
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
            mUserDesc = m.getString(TAG_USER_DESC);
        }
        mReference = m.getString(TAG_REFERENCE);
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (TAG_ADVANTAGE.equals(type) || TAG_ADVANTAGE_CONTAINER.equals(type)) {
                addChild(new Advantage(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, boolean forUndo) throws IOException {
        w.keyValueNot(ATTR_ROUND_COST_DOWN, mRoundCostDown, false);
        w.keyValueNot(ATTR_ALLOW_HALF_LEVELS, mAllowHalfLevels, false);
        w.keyValueNot(ATTR_DISABLED, mDisabled, false);
        if (canHaveChildren() && mContainerType != AdvantageContainerType.GROUP) {
            w.keyValue(KEY_CONTAINER_TYPE, Enums.toId(mContainerType));
        }
        w.keyValue(TAG_NAME, mName);
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
                w.keyValue(TAG_LEVELS, levels.toString());
            }
            w.keyValueNot(TAG_BASE_POINTS, mPoints, 0);
            w.keyValueNot(TAG_POINTS_PER_LEVEL, mPointsPerLevel, 0);
            WeaponStats.saveList(w, KEY_WEAPONS, mWeapons);
        }
        if (mCR != SelfControlRoll.NONE_REQUIRED) {
            w.keyValue(TAG_CR, mCR.getCR());
            if (mCRAdj != SelfControlRollAdjustments.NONE) {
                w.keyValue(KEY_CR_ADJ, Enums.toId(mCRAdj));
            }
        }
        saveList(w, KEY_MODIFIERS, mModifiers, false);
        if (getDataFile() instanceof GURPSCharacter) {
            w.keyValueNot(TAG_USER_DESC, mUserDesc, "");
        }
        w.keyValueNot(TAG_REFERENCE, mReference, "");
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
            notifySingle(ID_CONTAINER_TYPE);
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
            notifySingle(ID_TYPE);
            return true;
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Advantage");
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
            notifySingle(ID_USER_DESC);
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
            notifySingle(ID_NAME);
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
            notifySingle(ID_CR);
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
            notifySingle(ID_CR);
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
            notifySingle(ID_LEVELS);
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
            notifySingle(ID_HALF_LEVEL);
            return true;
        }
        return false;
    }

    /**
     * @param factor The number of levels or half levels to set.
     * @return Whether it was modified.
     */
    public boolean adjustLevel(int factor) {
        if (factor == 0) {
            return false;
        }
        if (!mAllowHalfLevels) {
            return setLevels(Math.max(mLevels + factor, 0));
        }
        int halfLevels = mLevels * 2 + (mHalfLevel ? 1 : 0) + factor;
        if (halfLevels < 0) {
            halfLevels = 0;
        }
        boolean modified = setHalfLevel((halfLevels & 1) == 1);
        modified |= setLevels(halfLevels / 2);
        return modified;
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
            startNotify();
            notify(ID_DISABLED, this);
            notify(ID_POINTS, this);
            endNotify();
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
                    switch (one.getAffects()) {
                    case TOTAL:
                    case BASE_ONLY:
                    default:
                        basePoints += modifier;
                        break;
                    case LEVELS_ONLY:
                        pointsPerLevel += modifier;
                        break;
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
            if (getDataFile().useMultiplicativeModifiers()) {
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
            notifySingle(ID_POINTS);
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
            notifySingle(ID_POINTS);
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
            notifySingle(ID_REFERENCE);
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
            notifySingle(ID_ROUND_COST_DOWN);
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
            notifySingle(ID_ALLOW_HALF_LEVELS);
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

    private static int getTypeFromText(String text) {
        int type = 0;
        if (text.contains(TYPE_MENTAL)) {
            type |= TYPE_MASK_MENTAL;
        }
        if (text.contains(TYPE_PHYSICAL)) {
            type |= TYPE_MASK_PHYSICAL;
        }
        if (text.contains(TYPE_SOCIAL)) {
            type |= TYPE_MASK_SOCIAL;
        }
        if (text.contains(TYPE_EXOTIC)) {
            type |= TYPE_MASK_EXOTIC;
        }
        if (text.contains(TYPE_SUPERNATURAL)) {
            type |= TYPE_MASK_SUPERNATURAL;
        }
        return type;
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
                if (buffer.length() > 0) {
                    buffer.append("/");
                }
                buffer.append(TYPE_PHYSICAL);
            }
            if ((type & TYPE_MASK_SOCIAL) != 0) {
                if (buffer.length() > 0) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_SOCIAL);
            }
            if ((type & TYPE_MASK_EXOTIC) != 0) {
                if (buffer.length() > 0) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_EXOTIC);
            }
            if ((type & TYPE_MASK_SUPERNATURAL) != 0) {
                if (buffer.length() > 0) {
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
            notifySingle(ID_WEAPON_STATUS_CHANGED);
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

    /**
     * @param modifiers The value to set for modifiers.
     * @return {@code true} if modifiers changed
     */
    public boolean setModifiers(List<? extends Modifier> modifiers) {
        List<AdvantageModifier> in = new FilteredList<>(modifiers, AdvantageModifier.class);
        if (!mModifiers.equals(in)) {
            mModifiers = in;
            notifySingle(ID_MODIFIER_STATUS_CHANGED);
            update();
            return true;
        }
        return false;
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
        if (builder.length() > 0) {
            // Remove the trailing MODIFIER_SEPARATOR
            builder.setLength(builder.length() - MODIFIER_SEPARATOR.length());
        }
        return builder.toString();
    }

    @Override
    protected String getCategoryID() {
        return ID_CATEGORY;
    }

    /** @return The "secondary" text, the text display below an Advantage. */
    @Override
    protected String getSecondaryText() {
        StringBuilder builder = new StringBuilder();
        if (getDataFile().userDescriptionDisplay().inline()) {
            String txt = getUserDesc();
            if (!txt.isBlank()) {
                builder.append(txt);
            }
        }
        String txt = super.getSecondaryText();
        if (!txt.isBlank()) {
            if (builder.length() > 0) {
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
