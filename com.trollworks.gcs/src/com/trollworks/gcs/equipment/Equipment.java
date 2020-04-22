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

package com.trollworks.gcs.equipment;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.collections.FilteredList;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.feature.ContainedWeightReduction;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.io.xml.XMLReader;
import com.trollworks.gcs.io.xml.XMLWriter;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponStats;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;

/** A piece of equipment. */
public class Equipment extends ListRow implements HasSourceReference {
    private static final int                     CURRENT_VERSION            = 6;
    private static final int                     EQUIPMENT_SPLIT_VERSION    = 6;
    private static final String                  DEFAULT_LEGALITY_CLASS     = "4";
    /** The extension for Equipment lists. */
    public static final  String                  OLD_EQUIPMENT_EXTENSION    = "eqp";
    /** The XML tag used for items. */
    public static final  String                  TAG_EQUIPMENT              = "equipment";
    /** The XML tag used for containers. */
    public static final  String                  TAG_EQUIPMENT_CONTAINER    = "equipment_container";
    private static final String                  ATTRIBUTE_EQUIPPED         = "equipped";
    private static final String                  TAG_QUANTITY               = "quantity";
    private static final String                  TAG_USES                   = "uses";
    private static final String                  TAG_MAX_USES               = "max_uses";
    private static final String                  TAG_DESCRIPTION            = "description";
    private static final String                  TAG_TECH_LEVEL             = "tech_level";
    private static final String                  TAG_LEGALITY_CLASS         = "legality_class";
    private static final String                  TAG_VALUE                  = "value";
    private static final String                  TAG_WEIGHT                 = "weight";
    private static final String                  TAG_REFERENCE              = "reference";
    /** The prefix used in front of all IDs for the equipment. */
    public static final  String                  PREFIX                     = GURPSCharacter.CHARACTER_PREFIX + "equipment.";
    /** The field ID for equipped/carried/not carried changes. */
    public static final  String                  ID_EQUIPPED                = PREFIX + "Equipped";
    /** The field ID for quantity changes. */
    public static final  String                  ID_QUANTITY                = PREFIX + "Quantity";
    /** The field ID for uses changes. */
    public static final  String                  ID_USES                    = PREFIX + "Uses";
    /** The field ID for max uses changes. */
    public static final  String                  ID_MAX_USES                = PREFIX + "MaxUses";
    /** The field ID for description changes. */
    public static final  String                  ID_DESCRIPTION             = PREFIX + "Description";
    /** The field ID for tech level changes. */
    public static final  String                  ID_TECH_LEVEL              = PREFIX + "TechLevel";
    /** The field ID for legality changes. */
    public static final  String                  ID_LEGALITY_CLASS          = PREFIX + "LegalityClass";
    /** The field ID for value changes. */
    public static final  String                  ID_VALUE                   = PREFIX + "Value";
    /** The field ID for weight changes. */
    public static final  String                  ID_WEIGHT                  = PREFIX + "Weight";
    /** The field ID for extended value changes */
    public static final  String                  ID_EXTENDED_VALUE          = PREFIX + "ExtendedValue";
    /** The field ID for extended weight changes */
    public static final  String                  ID_EXTENDED_WEIGHT         = PREFIX + "ExtendedWeight";
    /** The field ID for page reference changes. */
    public static final  String                  ID_REFERENCE               = PREFIX + "Reference";
    /** The field ID for when the categories change. */
    public static final  String                  ID_CATEGORY                = PREFIX + "Category";
    /** The field ID for when the row hierarchy changes. */
    public static final  String                  ID_LIST_CHANGED            = PREFIX + "ListChanged";
    /** The field ID for when the equipment becomes or stops being a weapon. */
    public static final  String                  ID_WEAPON_STATUS_CHANGED   = PREFIX + "WeaponStatus";
    /** The field ID for when the equipment gets Modifiers. */
    public static final  String                  ID_MODIFIER_STATUS_CHANGED = PREFIX + "Modifier";
    private              boolean                 mEquipped;
    private              int                     mQuantity;
    private              int                     mUses;
    private              int                     mMaxUses;
    private              String                  mDescription;
    private              String                  mTechLevel;
    private              String                  mLegalityClass;
    private              double                  mValue;
    private              WeightValue             mWeight;
    private              double                  mExtendedValue;
    private              WeightValue             mExtendedWeight;
    private              String                  mReference;
    private              List<WeaponStats>       mWeapons;
    private              List<EquipmentModifier> mModifiers;

    /**
     * Creates a new equipment.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    public Equipment(DataFile dataFile, boolean isContainer) {
        super(dataFile, isContainer);
        mEquipped = true;
        mQuantity = 1;
        mDescription = I18n.Text("Equipment");
        mTechLevel = "";
        mLegalityClass = DEFAULT_LEGALITY_CLASS;
        mReference = "";
        mWeight = new WeightValue(0, DisplayPreferences.getWeightUnits());
        mExtendedWeight = new WeightValue(mWeight);
        mWeapons = new ArrayList<>();
        mModifiers = new ArrayList<>();
    }

    /**
     * Creates a clone of an existing equipment and associates it with the specified data file.
     *
     * @param dataFile  The data file to associate it with.
     * @param equipment The equipment to clone.
     * @param deep      Whether or not to clone the children, grandchildren, etc.
     */
    public Equipment(DataFile dataFile, Equipment equipment, boolean deep) {
        super(dataFile, equipment);
        boolean forSheet = dataFile instanceof GURPSCharacter;
        mEquipped = !forSheet || equipment.mEquipped;
        mQuantity = forSheet ? equipment.mQuantity : 1;
        mUses = forSheet ? equipment.mUses : equipment.mMaxUses;
        mMaxUses = equipment.mMaxUses;
        mDescription = equipment.mDescription;
        mTechLevel = equipment.mTechLevel;
        mLegalityClass = equipment.mLegalityClass;
        mValue = equipment.mValue;
        mWeight = new WeightValue(equipment.mWeight);
        mReference = equipment.mReference;
        mWeapons = new ArrayList<>(equipment.mWeapons.size());
        for (WeaponStats weapon : equipment.mWeapons) {
            if (weapon instanceof MeleeWeaponStats) {
                mWeapons.add(new MeleeWeaponStats(this, (MeleeWeaponStats) weapon));
            } else if (weapon instanceof RangedWeaponStats) {
                mWeapons.add(new RangedWeaponStats(this, (RangedWeaponStats) weapon));
            }
        }
        mModifiers = new ArrayList<>(equipment.mModifiers.size());
        for (EquipmentModifier modifier : equipment.mModifiers) {
            mModifiers.add(new EquipmentModifier(mDataFile, modifier));
        }
        mExtendedValue = mQuantity * getAdjustedValue();
        mExtendedWeight = new WeightValue(getAdjustedWeight());
        mExtendedWeight.setValue(mExtendedWeight.getValue() * mQuantity);
        if (deep) {
            int count = equipment.getChildCount();
            for (int i = 0; i < count; i++) {
                addChild(new Equipment(dataFile, (Equipment) equipment.getChild(i), true));
            }
        }
    }

    /**
     * Loads an equipment and associates it with the specified data file.
     *
     * @param dataFile The data file to associate it with.
     * @param reader   The XML reader to load from.
     * @param state    The {@link LoadState} to use.
     */
    public Equipment(DataFile dataFile, XMLReader reader, LoadState state) throws IOException {
        this(dataFile, TAG_EQUIPMENT_CONTAINER.equals(reader.getName()));
        load(reader, state);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof Equipment && super.isEquivalentTo(obj)) {
            Equipment row = (Equipment) obj;
            if (mQuantity == row.mQuantity && mUses == row.mUses && mMaxUses == row.mMaxUses && mValue == row.mValue && mEquipped == row.mEquipped && mWeight.equals(row.mWeight) && mDescription.equals(row.mDescription) && mTechLevel.equals(row.mTechLevel) && mLegalityClass.equals(row.mLegalityClass) && mReference.equals(row.mReference)) {
                if (mWeapons.equals(row.mWeapons)) {
                    return mModifiers.equals(row.mModifiers);
                }
            }
        }
        return false;
    }

    @Override
    public String getLocalizedName() {
        return I18n.Text("Equipment");
    }

    @Override
    public String getListChangedID() {
        return ID_LIST_CHANGED;
    }

    @Override
    public String getXMLTagName() {
        return canHaveChildren() ? TAG_EQUIPMENT_CONTAINER : TAG_EQUIPMENT;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
    }

    @Override
    public String getRowType() {
        return I18n.Text("Equipment");
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mEquipped = true;
        mQuantity = 1;
        mUses = 0;
        mMaxUses = 0;
        mDescription = I18n.Text("Equipment");
        mTechLevel = "";
        mLegalityClass = DEFAULT_LEGALITY_CLASS;
        mReference = "";
        mValue = 0.0;
        mWeight.setValue(0.0);
        mWeapons = new ArrayList<>();
        mModifiers = new ArrayList<>();
    }

    @Override
    protected void loadAttributes(XMLReader reader, LoadState state) {
        super.loadAttributes(reader, state);
        if (mDataFile instanceof GURPSCharacter) {
            mEquipped = state.mDataItemVersion == 0 || state.mDataItemVersion >= EQUIPMENT_SPLIT_VERSION ? reader.isAttributeSet(ATTRIBUTE_EQUIPPED) : "equipped".equals(reader.getAttribute("state"));
            if (state.mDataFileVersion < GURPSCharacter.SEPARATED_EQUIPMENT_VERSION) {
                if (!mEquipped && !"carried".equals(reader.getAttribute("state"))) {
                    state.mUncarriedEquipment.add(this);
                }
            }
        }
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (TAG_DESCRIPTION.equals(name)) {
            mDescription = reader.readText().replace("\n", " ");
        } else if (TAG_TECH_LEVEL.equals(name)) {
            mTechLevel = reader.readText().replace("\n", " ");
        } else if (TAG_LEGALITY_CLASS.equals(name)) {
            mLegalityClass = reader.readText().replace("\n", " ");
        } else if (TAG_VALUE.equals(name)) {
            mValue = reader.readDouble(0.0);
        } else if (TAG_WEIGHT.equals(name)) {
            mWeight = WeightValue.extract(reader.readText(), false);
        } else if (TAG_REFERENCE.equals(name)) {
            mReference = reader.readText().replace("\n", " ");
        } else if (TAG_USES.equals(name)) {
            mUses = reader.readInteger(0);
        } else if (TAG_MAX_USES.equals(name)) {
            mMaxUses = reader.readInteger(0);
        } else if (!state.mForUndo && (TAG_EQUIPMENT.equals(name) || TAG_EQUIPMENT_CONTAINER.equals(name))) {
            addChild(new Equipment(mDataFile, reader, state));
        } else if (EquipmentModifier.TAG_MODIFIER.equals(name)) {
            mModifiers.add(new EquipmentModifier(getDataFile(), reader, state));
        } else if (MeleeWeaponStats.TAG_ROOT.equals(name)) {
            mWeapons.add(new MeleeWeaponStats(this, reader));
        } else if (RangedWeaponStats.TAG_ROOT.equals(name)) {
            mWeapons.add(new RangedWeaponStats(this, reader));
        } else if (!canHaveChildren()) {
            if (TAG_QUANTITY.equals(name)) {
                mQuantity = reader.readInteger(1);
            } else {
                super.loadSubElement(reader, state);
            }
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void finishedLoading(LoadState state) {
        if (mMaxUses < 0) {
            mMaxUses = 0;
        }
        if (mUses > mMaxUses) {
            mUses = mMaxUses;
        } else if (mUses < 0) {
            mUses = 0;
        }
        updateExtendedValue(false);
        updateExtendedWeight(false);
        super.finishedLoading(state);
    }

    @Override
    protected void saveAttributes(XMLWriter out, boolean forUndo) {
        if (mDataFile instanceof GURPSCharacter) {
            out.writeAttribute(ATTRIBUTE_EQUIPPED, mEquipped);
        }
    }

    @Override
    protected void saveSelf(XMLWriter out, boolean forUndo) {
        if (!canHaveChildren()) {
            out.simpleTag(TAG_QUANTITY, mQuantity);
        }
        out.simpleTagNotEmpty(TAG_DESCRIPTION, mDescription);
        out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
        out.simpleTagNotEmpty(TAG_LEGALITY_CLASS, mLegalityClass);
        out.simpleTag(TAG_VALUE, mValue);
        if (mWeight.getNormalizedValue() != 0) {
            out.simpleTag(TAG_WEIGHT, mWeight.toString(false));
        }
        out.simpleTagNotEmpty(TAG_REFERENCE, mReference);
        out.simpleTagNotZero(TAG_USES, mUses);
        out.simpleTagNotZero(TAG_MAX_USES, mMaxUses);
        for (WeaponStats weapon : mWeapons) {
            weapon.save(out);
        }
        for (EquipmentModifier modifier : mModifiers) {
            modifier.save(out, forUndo);
        }
    }

    @Override
    public void update() {
        updateExtendedValue(true);
        updateExtendedWeight(true);
    }

    public void updateNoNotify() {
        updateExtendedValue(false);
        updateExtendedWeight(false);
    }

    /** @return The quantity. */
    public int getQuantity() {
        return mQuantity;
    }

    /**
     * @param quantity The quantity to set.
     * @return Whether it was modified.
     */
    public boolean setQuantity(int quantity) {
        if (quantity != mQuantity) {
            mQuantity = quantity;
            startNotify();
            notify(ID_QUANTITY, this);
            updateContainingWeights(true);
            updateContainingValues(true);
            endNotify();
            return true;
        }
        return false;
    }

    /** @return The number of times this item can be used. */
    public int getUses() {
        return mUses;
    }

    /** @param uses The number of times this item can be used. */
    public boolean setUses(int uses) {
        if (uses > mMaxUses) {
            uses = mMaxUses;
        } else if (uses < 0) {
            uses = 0;
        }
        if (uses != mUses) {
            mUses = uses;
            notifySingle(ID_USES);
            return true;
        }
        return false;
    }

    /** @return The maximum number of times this item can be used. */
    public int getMaxUses() {
        return mMaxUses;
    }

    /** @param maxUses The maximum number of times this item can be used. */
    public boolean setMaxUses(int maxUses) {
        if (maxUses < 0) {
            maxUses = 0;
        }
        if (maxUses != mMaxUses) {
            boolean notifyUsesToo = false;
            mMaxUses = maxUses;
            if (mMaxUses > mUses) {
                mUses = mMaxUses;
                notifyUsesToo = true;
            }
            startNotify();
            notify(ID_MAX_USES, this);
            if (notifyUsesToo) {
                notify(ID_USES, this);
            }
            endNotify();
            return true;
        }
        return false;
    }

    /** @return The description. */
    public String getDescription() {
        return mDescription;
    }

    /**
     * @param description The description to set.
     * @return Whether it was modified.
     */
    public boolean setDescription(String description) {
        if (!mDescription.equals(description)) {
            mDescription = description;
            notifySingle(ID_DESCRIPTION);
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
     * @return Whether it was modified.
     */
    public boolean setTechLevel(String techLevel) {
        if (!mTechLevel.equals(techLevel)) {
            mTechLevel = techLevel;
            notifySingle(ID_TECH_LEVEL);
            return true;
        }
        return false;
    }

    /** @return The legality class. */
    public String getLegalityClass() {
        return mLegalityClass;
    }

    public String getDisplayLegalityClass() {
        String lc = getLegalityClass().trim();
        switch (lc) {
        case "0":
            return I18n.Text("LC0: Banned");
        case "1":
            return I18n.Text("LC1: Military");
        case "2":
            return I18n.Text("LC2: Restricted");
        case "3":
            return I18n.Text("LC3: Licensed");
        case "4":
            return I18n.Text("LC4: Open");
        default:
            return lc;
        }
    }

    /**
     * @param legalityClass The legality class to set.
     * @return Whether it was modified.
     */
    public boolean setLegalityClass(String legalityClass) {
        if (!mLegalityClass.equals(legalityClass)) {
            mLegalityClass = legalityClass;
            notifySingle(ID_LEGALITY_CLASS);
            return true;
        }
        return false;
    }

    /** @return The value after any cost adjustments. */
    public double getAdjustedValue() {
        return getValueAdjustedForModifiers(mValue, getModifiers());
    }

    /**
     * @param value     The base value to adjust.
     * @param modifiers The modifiers to apply.
     * @return The adjusted value.
     */
    public static double getValueAdjustedForModifiers(double value, List<EquipmentModifier> modifiers) {
        double multiplier      = 0;
        int    multiplierCount = 0;
        double costFactor      = 0;
        double finalMultiplier = 0;
        double finalAddition   = 0;
        for (EquipmentModifier modifier : modifiers) {
            if (modifier.isEnabled()) {
                double amt = modifier.getCostAdjAmount();
                switch (modifier.getCostAdjType()) {
                case BASE_ADDITION:
                    value += amt;
                    break;
                case BASE_MULTIPLIER:
                    multiplier += amt;
                    multiplierCount++;
                    break;
                case COST_FACTOR:
                    costFactor += amt;
                    break;
                case FINAL_MULTIPLIER:
                    finalMultiplier += amt;
                    break;
                case FINAL_ADDITION:
                    finalAddition += amt;
                    break;
                }
            }
        }
        if (multiplier != 0 && costFactor != 0) {
            // Has a mix of base cost multipliers and cost factors... so we need to convert the base
            // cost multipliers into cost factors.
            costFactor += multiplier - multiplierCount;
            multiplier = 0;
        }
        if (costFactor != 0) {
            if (costFactor < -0.8) {
                costFactor = -0.8;
            }
            value *= 1 + costFactor;
        } else if (multiplier != 0) {
            if (multiplier < 0.2) {
                multiplier = 0.2;
            }
            value *= multiplier;
        }
        if (finalMultiplier != 0) {
            if (finalMultiplier < 0.2) {
                finalMultiplier = 0.2;
            }
            value *= finalMultiplier;
        }
        value += finalAddition;
        if (value < 0) {
            value = 0;
        }
        return value;
    }

    /** @return The value. */
    public double getValue() {
        return mValue;
    }

    /**
     * @param value The value to set.
     * @return Whether it was modified.
     */
    public boolean setValue(double value) {
        if (value != mValue) {
            mValue = value;
            startNotify();
            notify(ID_VALUE, this);
            updateContainingValues(true);
            endNotify();
            return true;
        }
        return false;
    }

    /** @return The extended value. */
    public double getExtendedValue() {
        return mExtendedValue;
    }

    /** @return The weight after any adjustments. */
    public WeightValue getAdjustedWeight() {
        return getWeightAdjustedForModifiers(mWeight, getModifiers());
    }

    /**
     * @param value     The base value to adjust.
     * @param modifiers The modifiers to apply.
     * @return The adjusted value.
     */
    public static WeightValue getWeightAdjustedForModifiers(WeightValue value, List<EquipmentModifier> modifiers) {
        double      multiplier    = 1;
        WeightValue finalAddition = new WeightValue(0, value.getUnits());
        value = new WeightValue(value);
        for (EquipmentModifier modifier : modifiers) {
            if (modifier.isEnabled()) {
                switch (modifier.getWeightAdjType()) {
                case BASE_ADDITION:
                    value.add(modifier.getWeightAdjAddition());
                    break;
                case MULTIPLIER:
                    multiplier *= modifier.getWeightAdjMultiplier();
                    break;
                case FINAL_ADDITION:
                    finalAddition.add(modifier.getWeightAdjAddition());
                    break;
                }
            }
        }
        if (multiplier != 1) {
            value.setValue(value.getValue() * multiplier);
        }
        value.add(finalAddition);
        if (value.getValue() < 0) {
            value.setValue(0);
        }
        return value;
    }

    /** @return The weight. */
    public WeightValue getWeight() {
        return mWeight;
    }

    /**
     * @param weight The weight to set.
     * @return Whether it was modified.
     */
    public boolean setWeight(WeightValue weight) {
        if (!mWeight.equals(weight)) {
            mWeight = new WeightValue(weight);
            startNotify();
            notify(ID_WEIGHT, this);
            updateContainingWeights(true);
            endNotify();
            return true;
        }
        return false;
    }

    private boolean updateExtendedWeight(boolean okToNotify) {
        WeightValue saved = mExtendedWeight;
        int         count = getChildCount();
        WeightUnits units = mWeight.getUnits();
        mExtendedWeight = new WeightValue(getAdjustedWeight().getValue() * mQuantity, units);
        WeightValue contained = new WeightValue(0, units);
        for (int i = 0; i < count; i++) {
            Equipment   one    = (Equipment) getChild(i);
            WeightValue weight = one.mExtendedWeight;
            if (SheetPreferences.areGurpsMetricRulesUsed()) {
                weight = units.isMetric() ? GURPSCharacter.convertToGurpsMetric(weight) : GURPSCharacter.convertFromGurpsMetric(weight);
            }
            contained.add(weight);
        }
        int         percentage = 0;
        WeightValue reduction  = new WeightValue(0, units);
        for (Feature feature : getFeatures()) {
            if (feature instanceof ContainedWeightReduction) {
                ContainedWeightReduction cwr = (ContainedWeightReduction) feature;
                if (cwr.isPercentage()) {
                    percentage += cwr.getPercentageReduction();
                } else {
                    reduction.add(cwr.getAbsoluteReduction());
                }
            }
        }
        for (EquipmentModifier modifier : getModifiers()) {
            if (modifier.isEnabled()) {
                for (Feature feature : modifier.getFeatures()) {
                    if (feature instanceof ContainedWeightReduction) {
                        ContainedWeightReduction cwr = (ContainedWeightReduction) feature;
                        if (cwr.isPercentage()) {
                            percentage += cwr.getPercentageReduction();
                        } else {
                            reduction.add(cwr.getAbsoluteReduction());
                        }
                    }
                }
            }
        }
        if (percentage > 0) {
            if (percentage >= 100) {
                contained = new WeightValue(0, units);
            } else {
                contained.subtract(new WeightValue(contained.getValue() * percentage / 100, contained.getUnits()));
            }
        }
        contained.subtract(reduction);
        if (contained.getNormalizedValue() > 0) {
            mExtendedWeight.add(contained);
        }
        if (getParent() instanceof Equipment) {
            ((Equipment) getParent()).updateContainingWeights(okToNotify);
        }
        if (!saved.equals(mExtendedWeight)) {
            if (okToNotify) {
                notify(ID_EXTENDED_WEIGHT, this);
            }
            return true;
        }
        return false;
    }

    private void updateContainingWeights(boolean okToNotify) {
        Row parent = this;
        while (parent instanceof Equipment) {
            Equipment parentRow = (Equipment) parent;
            if (parentRow.updateExtendedWeight(okToNotify)) {
                parent = parentRow.getParent();
            } else {
                break;
            }
        }
    }

    private boolean updateExtendedValue(boolean okToNotify) {
        double savedValue = mExtendedValue;
        int    count      = getChildCount();
        mExtendedValue = mQuantity * getAdjustedValue();
        for (int i = 0; i < count; i++) {
            Equipment child = (Equipment) getChild(i);
            mExtendedValue += child.mExtendedValue;
        }
        if (getParent() instanceof Equipment) {
            ((Equipment) getParent()).updateContainingValues(okToNotify);
        }
        if (savedValue != mExtendedValue) {
            if (okToNotify) {
                notify(ID_EXTENDED_VALUE, this);
            }
            return true;
        }
        return false;
    }

    private void updateContainingValues(boolean okToNotify) {
        Row parent = this;
        while (parent instanceof Equipment) {
            Equipment parentRow = (Equipment) parent;
            if (parentRow.updateExtendedValue(okToNotify)) {
                parent = parentRow.getParent();
            } else {
                break;
            }
        }
    }

    /** @return The extended weight. */
    public WeightValue getExtendedWeight() {
        return mExtendedWeight;
    }

    /** @return Whether this item is equipped. */
    public boolean isEquipped() {
        return mEquipped;
    }

    /**
     * @param equipped The new equipped state.
     * @return Whether it was changed.
     */
    public boolean setEquipped(boolean equipped) {
        if (mEquipped != equipped) {
            mEquipped = equipped;
            notifySingle(ID_EQUIPPED);
            return true;
        }
        return false;
    }

    @Override
    public String getReference() {
        return mReference;
    }

    @Override
    public String getReferenceHighlight() {
        return getDescription();
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
    public boolean contains(String text, boolean lowerCaseOnly) {
        if (getDescription().toLowerCase().contains(text)) {
            return true;
        }
        return super.contains(text, lowerCaseOnly);
    }

    @Override
    public Object getData(Column column) {
        return EquipmentColumn.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return EquipmentColumn.values()[column.getID()].getDataAsText(this);
    }

    @Override
    public String toString() {
        return getDescription();
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
        return marker ? Images.EQP_MARKER : Images.EQP_FILE;
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new EquipmentEditor(this, getOwner().getProperty(EquipmentList.TAG_OTHER_ROOT) == null);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        super.fillWithNameableKeys(set);
        extractNameables(set, mDescription);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.fillWithNameableKeys(set);
            }
        }
        for (EquipmentModifier modifier : mModifiers) {
            modifier.fillWithNameableKeys(set);
        }
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        super.applyNameableKeys(map);
        mDescription = nameNameables(map, mDescription);
        for (WeaponStats weapon : mWeapons) {
            for (SkillDefault one : weapon.getDefaults()) {
                one.applyNameableKeys(map);
            }
        }
        for (EquipmentModifier modifier : mModifiers) {
            modifier.applyNameableKeys(map);
        }
    }

    /** @return The modifiers. */
    public List<EquipmentModifier> getModifiers() {
        return Collections.unmodifiableList(mModifiers);
    }

    /**
     * @param modifiers The value to set for modifiers.
     * @return {@code true} if modifiers changed
     */
    public boolean setModifiers(List<? extends Modifier> modifiers) {
        List<EquipmentModifier> in = new FilteredList<>(modifiers, EquipmentModifier.class);
        if (!mModifiers.equals(in)) {
            mModifiers = in;
            notifySingle(ID_MODIFIER_STATUS_CHANGED);
            return true;
        }
        return false;
    }

    /**
     * @param name The name to match against. Case-insensitive.
     * @return The first modifier that matches the name.
     */
    public EquipmentModifier getActiveModifierFor(String name) {
        for (EquipmentModifier m : getModifiers()) {
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
        for (EquipmentModifier modifier : mModifiers) {
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

    @Override
    public String getToolTip(Column column) {
        return EquipmentColumn.values()[column.getID()].getToolTip(this);
    }
}
