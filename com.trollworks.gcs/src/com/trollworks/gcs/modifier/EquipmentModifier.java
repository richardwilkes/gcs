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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.WeightValue;

import java.io.IOException;

public class EquipmentModifier extends Modifier {
    private static final int    CURRENT_JSON_VERSION   = 1;
    public static final  String KEY_MODIFIER           = "eqp_modifier";
    public static final  String KEY_MODIFIER_CONTAINER = "eqp_modifier_container";
    public static final  String KEY_COST_ADJ           = "cost";
    private static final String KEY_TECH_LEVEL         = "tech_level";
    public static final  String KEY_COST_TYPE          = "cost_type";
    public static final  String KEY_WEIGHT_ADJ         = "weight";
    public static final  String KEY_WEIGHT_TYPE        = "weight_type";

    private static final String DEFAULT_COST_AMOUNT = "+0";

    private EquipmentModifierCostType   mCostType;
    private String                      mCostAmount;
    private EquipmentModifierWeightType mWeightType;
    private String                      mWeightAmount;
    private String                      mTechLevel;

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file  The {@link DataFile} to use.
     * @param other Another {@link EquipmentModifier} to clone.
     * @param deep  Whether or not to clone the children, grandchildren, etc.
     */
    public EquipmentModifier(DataFile file, EquipmentModifier other, boolean deep) {
        super(file, other);
        mCostType = other.mCostType;
        mCostAmount = other.mCostAmount;
        mWeightType = other.mWeightType;
        mWeightAmount = other.mWeightAmount;
        mTechLevel = other.mTechLevel;
        if (deep) {
            int count = other.getChildCount();
            for (int i = 0; i < count; i++) {
                addChild(new EquipmentModifier(file, (EquipmentModifier) other.getChild(i), true));
            }
        }
    }

    public EquipmentModifier(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, KEY_MODIFIER_CONTAINER.equals(m.getString(DataFile.KEY_TYPE)));
        load(dataFile, m, state);
    }

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file        The {@link DataFile} to use.
     * @param isContainer Whether or not this row allows children.
     */
    public EquipmentModifier(DataFile file, boolean isContainer) {
        super(file, isContainer);
        mCostType = EquipmentModifierCostType.TO_ORIGINAL_COST;
        mCostAmount = DEFAULT_COST_AMOUNT;
        mWeightType = EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT;
        mWeightAmount = getDefaultWeightAmount();
        mTechLevel = "";
    }

    private String getDefaultWeightAmount() {
        return "+" + new WeightValue(Fixed6.ZERO, getDataFile().defaultWeightUnits());
    }

    @Override
    public EquipmentModifier cloneModifier(boolean deep) {
        return new EquipmentModifier(mDataFile, this, deep);
    }

    @Override
    public RetinaIcon getIcon(boolean marker) {
        return marker ? Images.EQM_MARKER : Images.EQM_FILE;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof EquipmentModifier && super.isEquivalentTo(obj)) {
            EquipmentModifier row = (EquipmentModifier) obj;
            return mCostType == row.mCostType && mCostAmount.equals(row.mCostAmount) && mWeightType == row.mWeightType && mWeightAmount.equals(row.mWeightAmount) && mTechLevel.equals(row.mTechLevel);
        }
        return false;
    }

    @Override
    public RowEditor<? extends ListRow> createEditor() {
        return new EquipmentModifierEditor(this);
    }

    /** @return The type of the cost modifier. */
    public EquipmentModifierCostType getCostAdjType() {
        return mCostType;
    }

    /**
     * @param costType The type of cost modifier to set.
     * @return {@code true} if a change was made.
     */
    public boolean setCostAdjType(EquipmentModifierCostType costType) {
        if (costType != mCostType) {
            mCostType = costType;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The amount for the cost modifier. */
    public String getCostAdjAmount() {
        return mCostAmount;
    }

    /**
     * @param amount The amount for the cost modifier to set.
     * @return {@code true} if a change was made.
     */
    public boolean setCostAdjAmount(String amount) {
        amount = mCostType.format(amount, false);
        if (!mCostAmount.equals(amount)) {
            mCostAmount = amount;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The type of the weight modifier. */
    public EquipmentModifierWeightType getWeightAdjType() {
        return mWeightType;
    }

    /**
     * @param weightType The type of weight modifier to set.
     * @return {@code true} if a change was made.
     */
    public boolean setWeightAdjType(EquipmentModifierWeightType weightType) {
        if (weightType != mWeightType) {
            mWeightType = weightType;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The amount for the weight multiplier. */
    public String getWeightAdjAmount() {
        return mWeightAmount;
    }

    /**
     * @param amount The amount for the weight modifier to set.
     * @return {@code true} if a change was made.
     */
    public boolean setWeightAdjAmount(String amount) {
        amount = mWeightType.format(amount, getDataFile().defaultWeightUnits(), false);
        if (!mWeightAmount.equals(amount)) {
            mWeightAmount = amount;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_MODIFIER_CONTAINER : KEY_MODIFIER;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
    }

    @Override
    public Object getData(Column column) {
        return EquipmentModifierColumnID.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return EquipmentModifierColumnID.values()[column.getID()].getDataAsText(this);
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mCostType = EquipmentModifierCostType.TO_ORIGINAL_COST;
        mCostAmount = DEFAULT_COST_AMOUNT;
        mWeightType = EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT;
        mWeightAmount = getDefaultWeightAmount();
        mTechLevel = "";
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        if (!canHaveChildren()) {
            if (m.has(KEY_COST_TYPE)) {
                mCostType = Enums.extract(m.getString(KEY_COST_TYPE), EquipmentModifierCostType.values(), EquipmentModifierCostType.TO_ORIGINAL_COST);
                mCostAmount = mCostType.format(m.getString(KEY_COST_ADJ), false);
            }
            if (m.has(KEY_WEIGHT_TYPE)) {
                mWeightType = Enums.extract(m.getString(KEY_WEIGHT_TYPE), EquipmentModifierWeightType.values(), EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT);
                mWeightAmount = mWeightType.format(m.getString(KEY_WEIGHT_ADJ), getDataFile().defaultWeightUnits(), false);
            }
            mTechLevel = m.getString(KEY_TECH_LEVEL);
        }
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (KEY_MODIFIER.equals(type) || KEY_MODIFIER_CONTAINER.equals(type)) {
                addChild(new EquipmentModifier(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        super.saveSelf(w, saveType);
        if (!canHaveChildren()) {
            if (mCostType != EquipmentModifierCostType.TO_ORIGINAL_COST || !mCostAmount.equals(DEFAULT_COST_AMOUNT)) {
                w.keyValue(KEY_COST_TYPE, Enums.toId(mCostType));
                w.keyValue(KEY_COST_ADJ, mCostAmount);
            }
            if (mWeightType != EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT || !mWeightAmount.equals(getDefaultWeightAmount())) {
                w.keyValue(KEY_WEIGHT_TYPE, Enums.toId(mWeightType));
                w.keyValue(KEY_WEIGHT_ADJ, mWeightAmount);
            }
            w.keyValueNot(KEY_TECH_LEVEL, mTechLevel, "");
        }
    }

    /** @return A full description of this {@link EquipmentModifier}. */
    public String getFullDescription() {
        StringBuilder builder = new StringBuilder();
        String        modNote = getNotes();
        builder.append(this);
        if (!modNote.isEmpty()) {
            builder.append(" (");
            builder.append(modNote);
            builder.append(')');
        }
        if (mDataFile instanceof GURPSCharacter && ((GURPSCharacter) mDataFile).getSettings().showEquipmentModifierAdj()) {
            String costDesc   = getCostDescription();
            String weightDesc = getWeightDescription();
            if (!costDesc.isEmpty() || !weightDesc.isEmpty()) {
                builder.append(" [");
                builder.append(costDesc);
                if (!weightDesc.isEmpty()) {
                    if (!costDesc.isEmpty()) {
                        builder.append("; ");
                    }
                    builder.append(weightDesc);
                }
                builder.append(']');
            }
        }
        return builder.toString();
    }

    /** @return The formatted cost adjustment. */
    public String getCostDescription() {
        if (canHaveChildren() || (mCostType == EquipmentModifierCostType.TO_ORIGINAL_COST && DEFAULT_COST_AMOUNT.equals(mCostAmount))) {
            return "";
        }
        return mCostType.format(mCostAmount, true) + " " + mCostType.toShortString();
    }

    /** @return The formatted weight adjustment. */
    public String getWeightDescription() {
        if (canHaveChildren() || (mWeightType == EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT && getDefaultWeightAmount().equals(mWeightAmount))) {
            return "";
        }
        return mWeightType.format(mWeightAmount, getDataFile().defaultWeightUnits(), true) + " " + mWeightType.toShortString();
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
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getToolTip(Column column) {
        return EquipmentModifierColumnID.values()[column.getID()].getToolTip(this);
    }
}
