/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.io.xml.XMLReader;
import com.trollworks.gcs.io.xml.XMLWriter;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.WeightValue;

import java.io.IOException;

public class EquipmentModifier extends Modifier {
    private static final int                         CURRENT_VERSION       = 1;
    /** The root tag. */
    public static final  String                      TAG_MODIFIER          = "eqp_modifier";
    /** The attribute for the cost. */
    public static final  String                      TAG_COST_ADJ          = "cost";
    /** The attribute for the cost type. */
    public static final  String                      ATTRIBUTE_COST_TYPE   = "cost_type";
    /** The attribute for the weight. */
    public static final  String                      TAG_WEIGHT_ADJ        = "weight";
    /** The attribute for the weight type. */
    public static final  String                      ATTRIBUTE_WEIGHT_TYPE = "weight_type";
    /** The notification prefix used. */
    public static final  String                      NOTIFICATION_PREFIX   = "eqpmod" + Notifier.SEPARATOR;
    /** The notification ID for enabled changes. */
    public static final  String                      ID_ENABLED            = NOTIFICATION_PREFIX + ATTRIBUTE_ENABLED;
    /** The notification ID for list changes. */
    public static final  String                      ID_LIST_CHANGED       = NOTIFICATION_PREFIX + "list_changed";
    /** The notification ID for cost adjustment changes. */
    public static final  String                      ID_COST_ADJ           = NOTIFICATION_PREFIX + TAG_COST_ADJ;
    /** The notification ID for weight adjustment changes. */
    public static final  String                      ID_WEIGHT_ADJ         = NOTIFICATION_PREFIX + TAG_WEIGHT_ADJ;
    private              EquipmentModifierCostType   mCostType;
    private              double                      mCostAmount;
    private              EquipmentModifierWeightType mWeightType;
    private              double                      mWeightMultiplier;
    private              WeightValue                 mWeightAddition;

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file  The {@link DataFile} to use.
     * @param other Another {@link EquipmentModifier} to clone.
     */
    public EquipmentModifier(DataFile file, EquipmentModifier other) {
        super(file, other);
        mCostType = other.mCostType;
        mCostAmount = other.mCostAmount;
        mWeightType = other.mWeightType;
        mWeightMultiplier = other.mWeightMultiplier;
        mWeightAddition = new WeightValue(other.mWeightAddition);
    }

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file   The {@link DataFile} to use.
     * @param reader The {@link XMLReader} to use.
     * @param state  The {@link LoadState} to use.
     */
    public EquipmentModifier(DataFile file, XMLReader reader, LoadState state) throws IOException {
        super(file, reader, state);
    }

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file The {@link DataFile} to use.
     */
    public EquipmentModifier(DataFile file) {
        super(file);
        mCostType = EquipmentModifierCostType.COST_FACTOR;
        mCostAmount = 1;
        mWeightType = EquipmentModifierWeightType.MULTIPLIER;
        mWeightMultiplier = 1;
        mWeightAddition = new WeightValue(0, DisplayPreferences.getWeightUnits());
    }

    @Override
    public EquipmentModifier cloneModifier() {
        return new EquipmentModifier(mDataFile, this);
    }

    @Override
    public String getNotificationPrefix() {
        return NOTIFICATION_PREFIX;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof EquipmentModifier && super.isEquivalentTo(obj)) {
            EquipmentModifier row = (EquipmentModifier) obj;
            return mCostType == row.mCostType && mCostAmount == row.mCostAmount;
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
            notifySingle(ID_COST_ADJ);
            return true;
        }
        return false;
    }

    /** @return The amount for the cost modifier. */
    public double getCostAdjAmount() {
        return mCostAmount;
    }

    /**
     * @param costAmount The amount for the cost modifier to set.
     * @return {@code true} if a change was made.
     */
    public boolean setCostAdjAmount(double costAmount) {
        if (costAmount != mCostAmount) {
            mCostAmount = costAmount;
            notifySingle(ID_COST_ADJ);
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
            notifySingle(ID_WEIGHT_ADJ);
            return true;
        }
        return false;
    }

    /**
     * @return The amount for the weight multiplier. Only valid if {@code getWeightAdjType() ==
     *         EquipmentModifierWeightType.MULTIPLIER}.
     */
    public double getWeightAdjMultiplier() {
        return mWeightMultiplier;
    }

    /**
     * @param multiplier The amount for the weight multiplier. Only valid if {@code
     *                   getWeightAdjType() == EquipmentModifierWeightType.MULTIPLIER}.
     * @return {@code true} if a change was made.
     */
    public boolean setWeightAdjMultiplier(double multiplier) {
        if (mWeightType == EquipmentModifierWeightType.MULTIPLIER && multiplier != mWeightMultiplier) {
            mWeightMultiplier = multiplier;
            notifySingle(ID_WEIGHT_ADJ);
            return true;
        }
        return false;
    }

    /**
     * @return The amount for the weight addition. Only valid if {@code getWeightAdjType() ==
     *         EquipmentModifierWeightType.ADDITION}.
     */
    public WeightValue getWeightAdjAddition() {
        return mWeightAddition;
    }

    /**
     * @param addition The amount for the weight addition. Only valid if {@code getWeightAdjType()
     *                 != EquipmentModifierWeightType.MULTIPLIER}.
     * @return {@code true} if a change was made.
     */
    public boolean setWeightAdjAddition(WeightValue addition) {
        if (mWeightType != EquipmentModifierWeightType.MULTIPLIER && !mWeightAddition.equals(addition)) {
            mWeightAddition = new WeightValue(addition);
            notifySingle(ID_WEIGHT_ADJ);
            return true;
        }
        return false;
    }

    @Override
    public String getXMLTagName() {
        return TAG_MODIFIER;
    }

    @Override
    public int getXMLTagVersion() {
        return CURRENT_VERSION;
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
        mCostType = EquipmentModifierCostType.COST_FACTOR;
        mCostAmount = 1;
        mWeightType = EquipmentModifierWeightType.MULTIPLIER;
        mWeightMultiplier = 1;
        mWeightAddition = new WeightValue(0, DisplayPreferences.getWeightUnits());
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (TAG_COST_ADJ.equals(name)) {
            mCostType = Enums.extract(reader.getAttribute(ATTRIBUTE_COST_TYPE), EquipmentModifierCostType.values(), EquipmentModifierCostType.COST_FACTOR);
            mCostAmount = reader.readDouble(1);
        } else if (TAG_WEIGHT_ADJ.equals(name)) {
            mWeightType = Enums.extract(reader.getAttribute(ATTRIBUTE_WEIGHT_TYPE), EquipmentModifierWeightType.values(), EquipmentModifierWeightType.MULTIPLIER);
            switch (mWeightType) {
            case BASE_ADDITION:
            case FINAL_ADDITION:
                mWeightAddition = WeightValue.extract(reader.readText(), false);
                break;
            case MULTIPLIER:
            default:
                mWeightMultiplier = reader.readDouble(1);
                break;
            }
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void saveSelf(XMLWriter out, boolean forUndo) {
        super.saveSelf(out, forUndo);
        out.simpleTagWithAttribute(TAG_COST_ADJ, mCostAmount, ATTRIBUTE_COST_TYPE, Enums.toId(mCostType));
        switch (mWeightType) {
        case BASE_ADDITION:
        case FINAL_ADDITION:
            if (mWeightAddition.getNormalizedValue() != 0) {
                out.simpleTagWithAttribute(TAG_WEIGHT_ADJ, mWeightAddition.toString(false), ATTRIBUTE_WEIGHT_TYPE, Enums.toId(mWeightType));
            }
            break;
        case MULTIPLIER:
        default:
            if (mWeightMultiplier != 1) {
                out.simpleTagWithAttribute(TAG_WEIGHT_ADJ, mWeightMultiplier, ATTRIBUTE_WEIGHT_TYPE, Enums.toId(mWeightType));
            }
            break;
        }
    }

    /** @return A full description of this {@link EquipmentModifier}. */
    public String getFullDescription() {
        StringBuilder builder = new StringBuilder();
        String        modNote = getNotes();
        builder.append(toString());
        if (!modNote.isEmpty()) {
            builder.append(" (");
            builder.append(modNote);
            builder.append(')');
        }
        return builder.toString();
    }

    /** @return The formatted cost adjustment. */
    public String getCostDescription() {
        return (mCostType == EquipmentModifierCostType.MULTIPLIER ? Numbers.format(mCostAmount) : Numbers.formatWithForcedSign(mCostAmount)) + " " + mCostType;
    }

    /** @return The formatted weight adjustment. */
    public String getWeightDescription() {
        StringBuilder builder = new StringBuilder();
        switch (mWeightType) {
        case BASE_ADDITION:
        case FINAL_ADDITION:
            if (mWeightAddition.getNormalizedValue() != 0) {
                String weight = mWeightAddition.toString();
                if (!weight.startsWith("-")) {
                    builder.append('+');
                }
                builder.append(weight);
                builder.append(' ');
                builder.append(mWeightType);
            }
            break;
        case MULTIPLIER:
        default:
            if (mWeightMultiplier != 1) {
                builder.append(Numbers.format(mWeightMultiplier));
                builder.append(' ');
                builder.append(mWeightType);
            }
            break;
        }
        return builder.toString();
    }
}
