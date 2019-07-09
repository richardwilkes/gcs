/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.HasSourceReference;
import com.trollworks.gcs.common.LoadState;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.OldWeapon;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.gcs.widgets.outline.Switchable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Enums;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A GURPS Advantage. */
public class Advantage extends ListRow implements HasSourceReference, Switchable {
    @Localize("Advantage")
    @Localize(locale = "de", value = "Vorteil")
    @Localize(locale = "ru", value = "Преимущество")
    @Localize(locale = "es", value = "Ventaja")
    private static String DEFAULT_NAME;

    static {
        Localization.initialize();
    }

    private static final int           CURRENT_VERSION            = 2;
    /** The XML tag used for items. */
    public static final String         TAG_ADVANTAGE              = "advantage"; //$NON-NLS-1$
    /** The XML tag used for containers. */
    public static final String         TAG_ADVANTAGE_CONTAINER    = "advantage_container"; //$NON-NLS-1$
    private static final String        TAG_REFERENCE              = "reference"; //$NON-NLS-1$
    private static final String        TAG_OLD_POINTS             = "points"; //$NON-NLS-1$
    private static final String        TAG_BASE_POINTS            = "base_points"; //$NON-NLS-1$
    private static final String        TAG_POINTS_PER_LEVEL       = "points_per_level"; //$NON-NLS-1$
    private static final String        TAG_LEVELS                 = "levels"; //$NON-NLS-1$
    private static final String        TAG_TYPE                   = "type"; //$NON-NLS-1$
    private static final String        TAG_NAME                   = "name"; //$NON-NLS-1$
    private static final String        TAG_CR                     = "cr"; //$NON-NLS-1$
    private static final String        TYPE_MENTAL                = "Mental"; //$NON-NLS-1$
    private static final String        TYPE_PHYSICAL              = "Physical"; //$NON-NLS-1$
    private static final String        TYPE_SOCIAL                = "Social"; //$NON-NLS-1$
    private static final String        TYPE_EXOTIC                = "Exotic"; //$NON-NLS-1$
    private static final String        TYPE_SUPERNATURAL          = "Supernatural"; //$NON-NLS-1$
    private static final String        ATTR_DISABLED              = "disabled"; //$NON-NLS-1$
    private static final String        ATTR_ROUND_COST_DOWN       = "round_down"; //$NON-NLS-1$
    private static final String        ATTR_ALLOW_HALF_LEVELS     = "allow_half_levels"; //$NON-NLS-1$
    private static final String        ATTR_HALF_LEVEL            = "half_level"; //$NON-NLS-1$
    /** The prefix used in front of all IDs for the advantages. */
    public static final String         PREFIX                     = GURPSCharacter.CHARACTER_PREFIX + "advantage."; //$NON-NLS-1$
    /** The field ID for type changes. */
    public static final String         ID_TYPE                    = PREFIX + "Type"; //$NON-NLS-1$
    /** The field ID for container type changes. */
    public static final String         ID_CONTAINER_TYPE          = PREFIX + "ContainerType"; //$NON-NLS-1$
    /** The field ID for name changes. */
    public static final String         ID_NAME                    = PREFIX + "Name"; //$NON-NLS-1$
    /** The field ID for CR changes. */
    public static final String         ID_CR                      = PREFIX + "CR"; //$NON-NLS-1$
    /** The field ID for level changes. */
    public static final String         ID_LEVELS                  = PREFIX + "Levels"; //$NON-NLS-1$
    /** The field ID for half level. */
    public static final String         ID_HALF_LEVEL              = PREFIX + "HalfLevel"; //$NON-NLS-1$
    /** The field ID for round cost down changes. */
    public static final String         ID_ROUND_COST_DOWN         = PREFIX + "RoundCostDown"; //$NON-NLS-1$
    /** The field ID for disabled changes. */
    public static final String         ID_DISABLED                = PREFIX + "Disabled"; //$NON-NLS-1$
    /** The field ID for allowing half levels. */
    public static final String         ID_ALLOW_HALF_LEVELS       = PREFIX + "AllowHalfLevels"; //$NON-NLS-1$
    /** The field ID for point changes. */
    public static final String         ID_POINTS                  = PREFIX + "Points"; //$NON-NLS-1$
    /** The field ID for page reference changes. */
    public static final String         ID_REFERENCE               = PREFIX + "Reference"; //$NON-NLS-1$
    /** The field ID for when the categories change. */
    public static final String         ID_CATEGORY                = PREFIX + "Category"; //$NON-NLS-1$
    /** The field ID for when the row hierarchy changes. */
    public static final String         ID_LIST_CHANGED            = PREFIX + "ListChanged"; //$NON-NLS-1$
    /** The field ID for when the advantage becomes or stops being a weapon. */
    public static final String         ID_WEAPON_STATUS_CHANGED   = PREFIX + "WeaponStatus"; //$NON-NLS-1$
    /** The field ID for when the advantage gets Modifiers. */
    public static final String         ID_MODIFIER_STATUS_CHANGED = PREFIX + "Modifier"; //$NON-NLS-1$
    /** The type mask for mental advantages. */
    public static final int            TYPE_MASK_MENTAL           = 1 << 0;
    /** The type mask for physical advantages. */
    public static final int            TYPE_MASK_PHYSICAL         = 1 << 1;
    /** The type mask for social advantages. */
    public static final int            TYPE_MASK_SOCIAL           = 1 << 2;
    /** The type mask for exotic advantages. */
    public static final int            TYPE_MASK_EXOTIC           = 1 << 3;
    /** The type mask for supernatural advantages. */
    public static final int            TYPE_MASK_SUPERNATURAL     = 1 << 4;
    private int                        mType;
    private String                     mName;
    private SelfControlRoll            mCR;
    private SelfControlRollAdjustments mCRAdj;
    private int                        mLevels;
    private boolean                    mAllowHalfLevels;
    private boolean                    mHalfLevel;
    private int                        mPoints;
    private int                        mPointsPerLevel;
    private String                     mReference;
    private String                     mOldPointsString;
    private AdvantageContainerType     mContainerType;
    private ArrayList<WeaponStats>     mWeapons;
    private ArrayList<Modifier>        mModifiers;
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
        mType          = TYPE_MASK_PHYSICAL;
        mName          = DEFAULT_NAME;
        mCR            = SelfControlRoll.NONE_REQUIRED;
        mCRAdj         = SelfControlRollAdjustments.NONE;
        mLevels        = -1;
        mReference     = ""; //$NON-NLS-1$
        mContainerType = AdvantageContainerType.GROUP;
        mWeapons       = new ArrayList<>();
        mModifiers     = new ArrayList<>();
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
        mType            = advantage.mType;
        mName            = advantage.mName;
        mCR              = advantage.mCR;
        mCRAdj           = advantage.mCRAdj;
        mLevels          = advantage.mLevels;
        mHalfLevel       = advantage.mHalfLevel;
        mAllowHalfLevels = advantage.mAllowHalfLevels;
        mPoints          = advantage.mPoints;
        mPointsPerLevel  = advantage.mPointsPerLevel;
        mRoundCostDown   = advantage.mRoundCostDown;
        mDisabled        = advantage.mDisabled;
        mReference       = advantage.mReference;
        mContainerType   = advantage.mContainerType;
        mWeapons         = new ArrayList<>(advantage.mWeapons.size());
        for (WeaponStats weapon : advantage.mWeapons) {
            if (weapon instanceof MeleeWeaponStats) {
                mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
            } else if (weapon instanceof RangedWeaponStats) {
                mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
            }
        }
        mModifiers = new ArrayList<>(advantage.mModifiers.size());
        for (Modifier modifier : advantage.mModifiers) {
            mModifiers.add(new Modifier(mDataFile, modifier));
        }
        if (deep) {
            int count = advantage.getChildCount();

            for (int i = 0; i < count; i++) {
                addChild(new Advantage(dataFile, (Advantage) advantage.getChild(i), true));
            }
        }
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
        return DEFAULT_NAME;
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mType            = TYPE_MASK_PHYSICAL;
        mName            = DEFAULT_NAME;
        mCR              = SelfControlRoll.NONE_REQUIRED;
        mCRAdj           = SelfControlRollAdjustments.NONE;
        mLevels          = -1;
        mHalfLevel       = false;
        mAllowHalfLevels = false;
        mReference       = ""; //$NON-NLS-1$
        mContainerType   = AdvantageContainerType.GROUP;
        mPoints          = 0;
        mPointsPerLevel  = 0;
        mRoundCostDown   = false;
        mDisabled        = false;
        mOldPointsString = null;
        mWeapons         = new ArrayList<>();
        mModifiers       = new ArrayList<>();
    }

    @Override
    protected void loadAttributes(XMLReader reader, LoadState state) {
        super.loadAttributes(reader, state);
        mRoundCostDown   = reader.isAttributeSet(ATTR_ROUND_COST_DOWN);
        mDisabled        = reader.isAttributeSet(ATTR_DISABLED);
        mAllowHalfLevels = reader.isAttributeSet(ATTR_ALLOW_HALF_LEVELS);
        if (canHaveChildren()) {
            mContainerType = Enums.extract(reader.getAttribute(TAG_TYPE), AdvantageContainerType.values(), AdvantageContainerType.GROUP);
        }
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();

        if (TAG_NAME.equals(name)) {
            mName = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
        } else if (TAG_CR.equals(name)) {
            mCRAdj = Enums.extract(reader.getAttribute(SelfControlRoll.ATTR_ADJUSTMENT), SelfControlRollAdjustments.values(), SelfControlRollAdjustments.NONE);
            mCR    = SelfControlRoll.get(reader.readText());
        } else if (TAG_REFERENCE.equals(name)) {
            mReference = reader.readText().replace("\n", " "); //$NON-NLS-1$ //$NON-NLS-2$
        } else if (!state.mForUndo && (TAG_ADVANTAGE.equals(name) || TAG_ADVANTAGE_CONTAINER.equals(name))) {
            addChild(new Advantage(mDataFile, reader, state));
        } else if (Modifier.TAG_MODIFIER.equals(name)) {
            mModifiers.add(new Modifier(getDataFile(), reader, state));
        } else if (!canHaveChildren()) {
            if (TAG_TYPE.equals(name)) {
                mType = getTypeFromText(reader.readText());
            } else if (TAG_LEVELS.equals(name)) {
                // Read the attribute first as next operation clears attribute map
                mHalfLevel = mAllowHalfLevels && reader.isAttributeSet(ATTR_HALF_LEVEL);
                mLevels    = reader.readInteger(-1);
            } else if (TAG_OLD_POINTS.equals(name)) {
                mOldPointsString = reader.readText();
            } else if (TAG_BASE_POINTS.equals(name)) {
                mPoints = reader.readInteger(0);
            } else if (TAG_POINTS_PER_LEVEL.equals(name)) {
                mPointsPerLevel = reader.readInteger(0);
            } else if (MeleeWeaponStats.TAG_ROOT.equals(name)) {
                mWeapons.add(new MeleeWeaponStats(this, reader));
            } else if (RangedWeaponStats.TAG_ROOT.equals(name)) {
                mWeapons.add(new RangedWeaponStats(this, reader));
            } else if (OldWeapon.TAG_ROOT.equals(name)) {
                state.mOldWeapons.put(this, new OldWeapon(reader));
            } else {
                super.loadSubElement(reader, state);
            }
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void finishedLoading(LoadState state) {
        if (mOldPointsString != null) {
            // All this is here solely to support loading old data files
            mOldPointsString = mOldPointsString.trim();
            int slash = mOldPointsString.indexOf('/');
            if (slash == -1) {
                mPoints         = getSimpleNumber(mOldPointsString);
                mPointsPerLevel = 0;
            } else {
                mPoints = getSimpleNumber(mOldPointsString.substring(0, slash));
                if (mOldPointsString.length() > slash) {
                    mPointsPerLevel = getSimpleNumber(mOldPointsString.substring(slash + 1));
                } else {
                    mPointsPerLevel = 0;
                }
                if (mPoints == 0) {
                    mPoints         = mPointsPerLevel;
                    mPointsPerLevel = 0;
                }
            }
            if (hasLevel() && mPointsPerLevel == 0) {
                mPointsPerLevel = mPoints;
                mPoints         = 0;
            }
            mOldPointsString = null;
        }
        OldWeapon oldWeapon = state.mOldWeapons.remove(this);
        if (oldWeapon != null) {
            mWeapons.addAll(oldWeapon.getWeapons(this));
        }
        // We no longer have defaults... that was solely for the weapons
        setDefaults(new ArrayList<SkillDefault>());
        super.finishedLoading(state);
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
    protected void saveAttributes(XMLWriter out, boolean forUndo) {
        super.saveAttributes(out, forUndo);
        if (mRoundCostDown) {
            out.writeAttribute(ATTR_ROUND_COST_DOWN, mRoundCostDown);
        }
        if (mAllowHalfLevels) {
            out.writeAttribute(ATTR_ALLOW_HALF_LEVELS, mAllowHalfLevels);
        }
        if (mDisabled) {
            out.writeAttribute(ATTR_DISABLED, mDisabled);
        }
        if (canHaveChildren() && mContainerType != AdvantageContainerType.GROUP) {
            out.writeAttribute(TAG_TYPE, Enums.toId(mContainerType));
        }
    }

    @Override
    public void saveSelf(XMLWriter out, boolean forUndo) {
        out.simpleTag(TAG_NAME, mName);
        if (!canHaveChildren()) {
            out.simpleTag(TAG_TYPE, getTypeAsText());
            if (mLevels != -1) {
                if (mAllowHalfLevels) {
                    out.simpleTagWithAttribute(TAG_LEVELS, mLevels, ATTR_HALF_LEVEL, mHalfLevel);
                } else {
                    out.simpleTag(TAG_LEVELS, mLevels);
                }
            }
            if (mPoints != 0) {
                out.simpleTag(TAG_BASE_POINTS, mPoints);
            }
            if (mPointsPerLevel != 0) {
                out.simpleTag(TAG_POINTS_PER_LEVEL, mPointsPerLevel);
            }

            for (WeaponStats weapon : mWeapons) {
                weapon.save(out);
            }
        }
        mCR.save(out, TAG_CR, mCRAdj);
        for (Modifier modifier : mModifiers) {
            modifier.save(out, forUndo);
        }
        out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
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
        return DEFAULT_NAME;
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
                ArrayList<Integer> values = new ArrayList<>();
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

    /** @return <code>true</code> if this {@link Advantage} is disabled. */
    public boolean isDisabled() {
        return !isEnabled();
    }

    /** @return <code>true</code> if this {@link Advantage} is enabled. */
    @Override
    public boolean isEnabled() {
        if (mDisabled) {
            return false;
        }
        Row parent = getParent();
        if (parent != null && parent instanceof Advantage) {
            return ((Advantage) parent).isEnabled();
        }
        return true;
    }

    /**
     * @return <code>true</code> if this {@link Advantage} is enabled, regardless of parent nodes.
     */
    public boolean isSelfEnabled() {
        return !mDisabled;
    }

    /**
     * @param enabled Whether the {@link Advantage} should be enabled or not.
     * @return Whether it was modified.
     */
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
     * @param modifiers      The {@link Modifier}s to apply.
     * @param roundCostDown  Whether the point cost should be rounded down rather than up, as is
     *                       normal for most GURPS rules.
     * @return The total points, taking levels and modifiers into account.
     */
    public static int getAdjustedPoints(int basePoints, int levels, boolean halfLevel, int pointsPerLevel, SelfControlRoll cr, Collection<Modifier> modifiers, boolean roundCostDown) {
        int    baseEnh    = 0;
        int    levelEnh   = 0;
        int    baseLim    = 0;
        int    levelLim   = 0;
        double multiplier = cr.getMultiplier();

        for (Modifier one : modifiers) {
            if (one.isEnabled()) {
                int modifier = one.getCostModifier();
                switch (one.getCostType()) {
                case PERCENTAGE:
                default:
                    switch (one.getAffects()) {
                    case TOTAL:
                    default:
                        if (modifier < 0) { // Limitation
                            baseLim  += modifier;
                            levelLim += modifier;
                        } else { // Enhancement
                            baseEnh  += modifier;
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
            int baseMod  = 0;
            int levelMod = 0;

            if (SheetPreferences.areOptionalModifierRulesUsed()) {
                if (baseEnh == levelEnh && baseLim == levelLim) {
                    modifiedBasePoints = modifyPoints(basePoints + leveledPoints, baseEnh);
                    modifiedBasePoints = modifyPoints(modifiedBasePoints, Math.max(baseLim, -80));
                } else {
                    modifiedBasePoints  = modifyPoints(basePoints, baseEnh);
                    modifiedBasePoints  = modifyPoints(modifiedBasePoints, Math.max(baseLim, -80));
                    leveledPoints       = modifyPoints(leveledPoints, levelEnh);
                    leveledPoints       = modifyPoints(leveledPoints, Math.max(levelLim, -80));
                    basePoints         += leveledPoints;
                }
            } else {
                baseMod  = Math.max(baseEnh + baseLim, -80);
                levelMod = Math.max(levelEnh + levelLim, -80);
                if (baseMod == levelMod) {
                    modifiedBasePoints = modifyPoints(basePoints + leveledPoints, baseMod);
                } else {
                    modifiedBasePoints = modifyPoints(basePoints, baseMod) + modifyPoints(leveledPoints, levelMod);
                }
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

    private static int getSimpleNumber(String buffer) {
        try {
            return Integer.parseInt(buffer);
        } catch (Exception exception) {
            return 0;
        }
    }

    @Override
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getName().toLowerCase().indexOf(text) != -1) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    private static int getTypeFromText(String text) {
        int type = 0;

        if (text.indexOf(TYPE_MENTAL) != -1) {
            type |= TYPE_MASK_MENTAL;
        }
        if (text.indexOf(TYPE_PHYSICAL) != -1) {
            type |= TYPE_MASK_PHYSICAL;
        }
        if (text.indexOf(TYPE_SOCIAL) != -1) {
            type |= TYPE_MASK_SOCIAL;
        }
        if (text.indexOf(TYPE_EXOTIC) != -1) {
            type |= TYPE_MASK_EXOTIC;
        }
        if (text.indexOf(TYPE_SUPERNATURAL) != -1) {
            type |= TYPE_MASK_SUPERNATURAL;
        }
        return type;
    }

    /** @return The type as a text string. */
    public String getTypeAsText() {
        if (!canHaveChildren()) {
            String        separator = ", "; //$NON-NLS-1$
            StringBuilder buffer    = new StringBuilder();
            int           type      = getType();

            if ((type & Advantage.TYPE_MASK_MENTAL) != 0) {
                buffer.append(TYPE_MENTAL);
            }
            if ((type & Advantage.TYPE_MASK_PHYSICAL) != 0) {
                if (buffer.length() > 0) {
                    buffer.append("/"); //$NON-NLS-1$
                }
                buffer.append(TYPE_PHYSICAL);
            }
            if ((type & Advantage.TYPE_MASK_SOCIAL) != 0) {
                if (buffer.length() > 0) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_SOCIAL);
            }
            if ((type & Advantage.TYPE_MASK_EXOTIC) != 0) {
                if (buffer.length() > 0) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_EXOTIC);
            }
            if ((type & Advantage.TYPE_MASK_SUPERNATURAL) != 0) {
                if (buffer.length() > 0) {
                    buffer.append(separator);
                }
                buffer.append(TYPE_SUPERNATURAL);
            }
            return buffer.toString();
        }
        return ""; //$NON-NLS-1$
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
                    builder.append('\u00bd');
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
    public StdImage getIcon(boolean large) {
        return GCSImages.getAdvantagesIcons().getImage(large ? 64 : 16);
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new AdvantageEditor(this);
    }

    @Override
    public void fillWithNameableKeys(HashSet<String> set) {
        super.fillWithNameableKeys(set);
        extractNameables(set, mName);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.fillWithNameableKeys(set);
            }
        }
        for (Modifier modifier : mModifiers) {
            modifier.fillWithNameableKeys(set);
        }
    }

    @Override
    public void applyNameableKeys(HashMap<String, String> map) {
        super.applyNameableKeys(map);
        mName = nameNameables(map, mName);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.applyNameableKeys(map);
            }
        }
        for (Modifier modifier : mModifiers) {
            modifier.applyNameableKeys(map);
        }
    }

    /** @return The modifiers. */
    public List<Modifier> getModifiers() {
        return Collections.unmodifiableList(mModifiers);
    }

    /** @return The modifiers including those inherited from parent row. */
    public List<Modifier> getAllModifiers() {
        ArrayList<Modifier> allModifiers = new ArrayList<>(mModifiers);
        if (getParent() != null) {
            allModifiers.addAll(((Advantage) getParent()).getAllModifiers());
        }
        return Collections.unmodifiableList(allModifiers);
    }

    /**
     * @param modifiers The value to set for modifiers.
     * @return {@code true} if modifiers changed
     */
    public boolean setModifiers(List<Modifier> modifiers) {
        if (!mModifiers.equals(modifiers)) {
            mModifiers = new ArrayList<>(modifiers);
            notifySingle(ID_MODIFIER_STATUS_CHANGED);
            return true;
        }
        return false;
    }

    /**
     * @param name The name to match against. Case-insensitive.
     * @return The first modifier that matches the name.
     */
    public Modifier getActiveModifierFor(String name) {
        for (Modifier m : getModifiers()) {
            if (m.isEnabled() && m.getName().equalsIgnoreCase(name)) {
                return m;
            }
        }
        return null;
    }

    private static final String MODIFIER_SEPARATOR = "; "; //$NON-NLS-1$

    @Override
    public String getModifierNotes() {
        StringBuilder builder = new StringBuilder();
        if (mCR != SelfControlRoll.NONE_REQUIRED) {
            builder.append(mCR);
            if (mCRAdj != SelfControlRollAdjustments.NONE) {
                builder.append(", "); //$NON-NLS-1$
                builder.append(mCRAdj.getDescription(getCR()));
            }
            builder.append(MODIFIER_SEPARATOR);
        }
        for (Modifier modifier : mModifiers) {
            if (modifier.isEnabled()) {
                builder.append(modifier.getFullDescription());
                builder.append(MODIFIER_SEPARATOR);
            }
        }
        if (builder.length() > 0) {
            // Remove the trailing MODIFIER_SEPARATOR
            builder.setLength(builder.length() - MODIFIER_SEPARATOR.length());
            builder.append('.');
        }
        return builder.toString();
    }

    @Override
    protected String getCategoryID() {
        return ID_CATEGORY;
    }

    @Override
    public boolean alwaysShowToolTip(Column column) {
        return AdvantageColumn.values()[column.getID()].showToolTip();
    }

    @Override
    public String getToolTip(Column column) {
        return AdvantageColumn.values()[column.getID()].getToolTip(this);
    }
}
