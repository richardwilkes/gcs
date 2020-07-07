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
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.feature.ContainedWeightReduction;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.menu.item.HasSourceReference;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.modifier.EquipmentModifierCostType;
import com.trollworks.gcs.modifier.EquipmentModifierWeightType;
import com.trollworks.gcs.modifier.Fraction;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.modifier.ModifierCostValueType;
import com.trollworks.gcs.modifier.ModifierWeightValueType;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FilteredList;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.utility.xml.XMLReader;
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
    private static final int                     CURRENT_JSON_VERSION       = 1;
    private static final int                     CURRENT_VERSION            = 7;
    private static final int                     EQUIPMENT_SPLIT_VERSION    = 6;
    private static final String                  DEFAULT_LEGALITY_CLASS     = "4";
    /** The XML tag used for items. */
    public static final  String                  TAG_EQUIPMENT              = "equipment";
    /** The XML tag used for containers. */
    public static final  String                  TAG_EQUIPMENT_CONTAINER    = "equipment_container";
    private static final String                  KEY_WEAPONS                = "weapons";
    private static final String                  KEY_MODIFIERS              = "modifiers";
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
    private static final Fixed6                  MIN_CF                     = new Fixed6("-0.8", Fixed6.ZERO, false);
    private              boolean                 mEquipped;
    private              int                     mQuantity;
    private              int                     mUses;
    private              int                     mMaxUses;
    private              String                  mDescription;
    private              String                  mTechLevel;
    private              String                  mLegalityClass;
    private              Fixed6                  mValue;
    private              WeightValue             mWeight;
    private              Fixed6                  mExtendedValue;
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
        mValue = Fixed6.ZERO;
        mExtendedValue = Fixed6.ZERO;
        mWeight = new WeightValue(Fixed6.ZERO, dataFile.defaultWeightUnits());
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
            mModifiers.add(new EquipmentModifier(mDataFile, modifier, false));
        }
        mExtendedValue = new Fixed6(mQuantity).mul(getAdjustedValue());
        mExtendedWeight = new WeightValue(getAdjustedWeight());
        mExtendedWeight.setValue(mExtendedWeight.getValue().mul(new Fixed6(mQuantity)));
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
     * @param m        The {@JsonMap} to load from.
     * @param state    The {@link LoadState} to use.
     */
    public Equipment(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, m.getString(DataFile.KEY_TYPE).equals(TAG_EQUIPMENT_CONTAINER));
        load(m, state);
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
            if (mQuantity == row.mQuantity && mUses == row.mUses && mMaxUses == row.mMaxUses && mValue.equals(row.mValue) && mEquipped == row.mEquipped && mWeight.equals(row.mWeight) && mDescription.equals(row.mDescription) && mTechLevel.equals(row.mTechLevel) && mLegalityClass.equals(row.mLegalityClass) && mReference.equals(row.mReference)) {
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
    public String getJSONTypeName() {
        return canHaveChildren() ? TAG_EQUIPMENT_CONTAINER : TAG_EQUIPMENT;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
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
        mValue = Fixed6.ZERO;
        mWeight.setValue(Fixed6.ZERO);
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
            mValue = new Fixed6(reader.readText(), Fixed6.ZERO, false);
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
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        if (mDataFile instanceof GURPSCharacter) {
            mEquipped = m.getBoolean(ATTRIBUTE_EQUIPPED);
        }
        if (!canHaveChildren()) {
            mQuantity = m.getInt(TAG_QUANTITY);
        }
        mDescription = m.getString(TAG_DESCRIPTION);
        mTechLevel = m.getString(TAG_TECH_LEVEL);
        mLegalityClass = m.getStringWithDefault(TAG_LEGALITY_CLASS, DEFAULT_LEGALITY_CLASS);
        mValue = new Fixed6(m.getString(TAG_VALUE), Fixed6.ZERO, false);
        mWeight = WeightValue.extract(m.getString(TAG_WEIGHT), false);
        mReference = m.getString(TAG_REFERENCE);
        mUses = m.getInt(TAG_USES);
        mMaxUses = m.getInt(TAG_MAX_USES);
        if (m.has(KEY_WEAPONS)) {
            WeaponStats.loadFromJSONArray(this, m.getArray(KEY_WEAPONS), mWeapons);
        }
        if (m.has(KEY_MODIFIERS)) {
            JsonArray a     = m.getArray(KEY_MODIFIERS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mModifiers.add(new EquipmentModifier(getDataFile(), a.getMap(i), state));
            }
        }
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (TAG_EQUIPMENT.equals(type) || TAG_EQUIPMENT_CONTAINER.equals(type)) {
                addChild(new Equipment(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, boolean forUndo) throws IOException {
        if (mDataFile instanceof GURPSCharacter) {
            w.keyValue(ATTRIBUTE_EQUIPPED, mEquipped);
        }
        if (!canHaveChildren()) {
            w.keyValueNot(TAG_QUANTITY, mQuantity, 0);
        }
        w.keyValueNot(TAG_DESCRIPTION, mDescription, "");
        w.keyValueNot(TAG_TECH_LEVEL, mTechLevel, "");
        w.keyValueNot(TAG_LEGALITY_CLASS, mLegalityClass, DEFAULT_LEGALITY_CLASS);
        if (!mValue.equals(Fixed6.ZERO)) {
            w.keyValue(TAG_VALUE, mValue.toString());
        }
        if (!mWeight.getNormalizedValue().equals(Fixed6.ZERO)) {
            w.keyValue(TAG_WEIGHT, mWeight.toString(false));
        }
        w.keyValueNot(TAG_REFERENCE, mReference, "");
        w.keyValueNot(TAG_USES, mUses, 0);
        w.keyValueNot(TAG_MAX_USES, mMaxUses, 0);
        WeaponStats.saveList(w, KEY_WEAPONS, mWeapons);
        saveList(w, KEY_MODIFIERS, mModifiers, false);
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
    public Fixed6 getAdjustedValue() {
        return getValueAdjustedForModifiers(mValue, getModifiers());
    }

    /**
     * @param value     The base value to adjust.
     * @param modifiers The modifiers to apply.
     * @return The adjusted value.
     */
    public static Fixed6 getValueAdjustedForModifiers(Fixed6 value, List<EquipmentModifier> modifiers) {
        // Apply all EquipmentModifierCostType.TO_ORIGINAL_COST
        Fixed6 cost = processNonCFStep(EquipmentModifierCostType.TO_ORIGINAL_COST, value, modifiers);

        // Apply all EquipmentModifierCostType.TO_BASE_COST
        Fixed6 cf    = Fixed6.ZERO;
        int    count = 0;
        for (EquipmentModifier modifier : modifiers) {
            if (modifier.isEnabled() && modifier.getCostAdjType() == EquipmentModifierCostType.TO_BASE_COST) {
                String                adj = modifier.getCostAdjAmount();
                ModifierCostValueType mvt = EquipmentModifierCostType.TO_BASE_COST.determineType(adj);
                Fixed6                amt = mvt.extractValue(adj, false);
                if (mvt == ModifierCostValueType.MULTIPLIER) {
                    amt = amt.sub(Fixed6.ONE);
                }
                cf = cf.add(amt);
                count++;
            }
        }
        if (!cf.equals(Fixed6.ZERO)) {
            if (cf.lessThan(MIN_CF)) {
                cf = MIN_CF;
            }
            cost = cost.mul(cf.add(Fixed6.ONE));
        }

        // Apply all EquipmentModifierCostType.TO_FINAL_BASE_COST
        cost = processNonCFStep(EquipmentModifierCostType.TO_FINAL_BASE_COST, cost, modifiers);

        // Apply all EquipmentModifierCostType.TO_FINAL_COST
        cost = processNonCFStep(EquipmentModifierCostType.TO_FINAL_COST, cost, modifiers);
        return cost.greaterThanOrEqual(Fixed6.ZERO) ? cost : Fixed6.ZERO;
    }

    private static Fixed6 processNonCFStep(EquipmentModifierCostType costType, Fixed6 value, List<EquipmentModifier> modifiers) {
        Fixed6 percentages = Fixed6.ZERO;
        Fixed6 additions   = Fixed6.ZERO;
        Fixed6 cost        = value;
        for (EquipmentModifier modifier : modifiers) {
            if (modifier.isEnabled() && modifier.getCostAdjType() == costType) {
                String                adj = modifier.getCostAdjAmount();
                ModifierCostValueType mvt = costType.determineType(adj);
                Fixed6                amt = mvt.extractValue(adj, false);
                switch (mvt) {
                case ADDITION:
                    additions = additions.add(amt);
                    break;
                case PERCENTAGE:
                    percentages = percentages.add(amt);
                    break;
                case MULTIPLIER:
                    cost = cost.mul(amt);
                    break;
                }
            }
        }
        cost = cost.add(additions);
        if (!percentages.equals(Fixed6.ZERO)) {
            cost = cost.add(value.mul(percentages.div(new Fixed6(100))));
        }
        return cost;
    }

    /** @return The value. */
    public Fixed6 getValue() {
        return mValue;
    }

    /**
     * @param value The value to set.
     * @return Whether it was modified.
     */
    public boolean setValue(Fixed6 value) {
        if (!mValue.equals(value)) {
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
    public Fixed6 getExtendedValue() {
        return mExtendedValue;
    }

    /** @return The weight after any adjustments. */
    public WeightValue getAdjustedWeight() {
        return getWeightAdjustedForModifiers(mWeight, getModifiers());
    }

    /**
     * @param weight    The base weight to adjust.
     * @param modifiers The modifiers to apply.
     * @return The adjusted value.
     */
    public WeightValue getWeightAdjustedForModifiers(WeightValue weight, List<EquipmentModifier> modifiers) {
        WeightUnits defUnits = getDataFile().defaultWeightUnits();
        weight = new WeightValue(weight);

        // Apply all EquipmentModifierWeightType.TO_ORIGINAL_COST
        Fixed6      percentages = Fixed6.ZERO;
        WeightValue original    = new WeightValue(weight);
        for (EquipmentModifier modifier : modifiers) {
            if (modifier.isEnabled() && modifier.getWeightAdjType() == EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT) {
                String                  adj = modifier.getWeightAdjAmount();
                ModifierWeightValueType mvt = EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT.determineType(adj);
                Fixed6                  amt = mvt.extractFraction(adj, false).value();
                if (mvt == ModifierWeightValueType.ADDITION) {
                    weight.add(new WeightValue(amt, ModifierWeightValueType.extractUnits(adj, defUnits)));
                } else {
                    percentages = percentages.add(amt);
                }
            }
        }
        if (!percentages.equals(Fixed6.ZERO)) {
            original.setValue(original.getValue().mul(percentages.div(new Fixed6(100))));
            weight.add(original);
        }

        // Apply all EquipmentModifierWeightType.TO_BASE_COST
        weight = processMultiplyAddWeightStep(EquipmentModifierWeightType.TO_BASE_WEIGHT, weight, defUnits, modifiers);

        // Apply all EquipmentModifierWeightType.TO_FINAL_BASE_COST
        weight = processMultiplyAddWeightStep(EquipmentModifierWeightType.TO_FINAL_BASE_WEIGHT, weight, defUnits, modifiers);

        // Apply all EquipmentModifierWeightType.TO_FINAL_COST
        weight = processMultiplyAddWeightStep(EquipmentModifierWeightType.TO_FINAL_WEIGHT, weight, defUnits, modifiers);
        if (weight.getValue().lessThan(Fixed6.ZERO)) {
            weight.setValue(Fixed6.ZERO);
        }
        return weight;
    }

    private WeightValue processMultiplyAddWeightStep(EquipmentModifierWeightType weightType, WeightValue weight, WeightUnits defUnits, List<EquipmentModifier> modifiers) {
        weight = new WeightValue(weight);
        WeightValue sum = new WeightValue(Fixed6.ZERO, weight.getUnits());
        for (EquipmentModifier modifier : modifiers) {
            if (modifier.isEnabled() && modifier.getWeightAdjType() == weightType) {
                String                  adj      = modifier.getWeightAdjAmount();
                ModifierWeightValueType mvt      = weightType.determineType(adj);
                Fraction                fraction = mvt.extractFraction(adj, false);
                switch (mvt) {
                case MULTIPLIER:
                    weight.setValue(weight.getValue().mul(fraction.mNumerator).div(fraction.mDenominator));
                    break;
                case PERCENTAGE_MULTIPLIER:
                    weight.setValue(weight.getValue().mul(fraction.mNumerator).div(fraction.mDenominator.mul(new Fixed6(100))));
                    break;
                case ADDITION:
                    sum.add(new WeightValue(fraction.value(), ModifierWeightValueType.extractUnits(adj, defUnits)));
                    break;
                default:
                    break;
                }
            }
        }
        weight.add(sum);
        return weight;
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
        mExtendedWeight = new WeightValue(getAdjustedWeight().getValue().mul(new Fixed6(mQuantity)), units);
        WeightValue contained = new WeightValue(Fixed6.ZERO, units);
        for (int i = 0; i < count; i++) {
            Equipment   one    = (Equipment) getChild(i);
            WeightValue weight = one.mExtendedWeight;
            if (mDataFile.useSimpleMetricConversions()) {
                weight = units.isMetric() ? GURPSCharacter.convertToGurpsMetric(weight) : GURPSCharacter.convertFromGurpsMetric(weight);
            }
            contained.add(weight);
        }
        Fixed6      percentage = Fixed6.ZERO;
        WeightValue reduction  = new WeightValue(Fixed6.ZERO, units);
        for (Feature feature : getFeatures()) {
            if (feature instanceof ContainedWeightReduction) {
                ContainedWeightReduction cwr = (ContainedWeightReduction) feature;
                if (cwr.isPercentage()) {
                    percentage = percentage.add(new Fixed6(cwr.getPercentageReduction()));
                } else {
                    reduction.add(cwr.getAbsoluteReduction(mDataFile.defaultWeightUnits()));
                }
            }
        }
        for (EquipmentModifier modifier : getModifiers()) {
            if (modifier.isEnabled()) {
                for (Feature feature : modifier.getFeatures()) {
                    if (feature instanceof ContainedWeightReduction) {
                        ContainedWeightReduction cwr = (ContainedWeightReduction) feature;
                        if (cwr.isPercentage()) {
                            percentage = percentage.add(new Fixed6(cwr.getPercentageReduction()));
                        } else {
                            reduction.add(cwr.getAbsoluteReduction(mDataFile.defaultWeightUnits()));
                        }
                    }
                }
            }
        }
        if (percentage.greaterThan(Fixed6.ZERO)) {
            Fixed6 oneHundred = new Fixed6(100);
            if (percentage.greaterThanOrEqual(oneHundred)) {
                contained = new WeightValue(Fixed6.ZERO, units);
            } else {
                contained.subtract(new WeightValue(contained.getValue().mul(percentage).div(oneHundred), contained.getUnits()));
            }
        }
        contained.subtract(reduction);
        if (contained.getNormalizedValue().greaterThan(Fixed6.ZERO)) {
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
        Fixed6 savedValue = mExtendedValue;
        int    count      = getChildCount();
        mExtendedValue = new Fixed6(mQuantity).mul(getAdjustedValue());
        for (int i = 0; i < count; i++) {
            Equipment child = (Equipment) getChild(i);
            mExtendedValue = mExtendedValue.add(child.mExtendedValue);
        }
        if (getParent() instanceof Equipment) {
            ((Equipment) getParent()).updateContainingValues(okToNotify);
        }
        if (!mExtendedValue.equals(savedValue)) {
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
            update();
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
