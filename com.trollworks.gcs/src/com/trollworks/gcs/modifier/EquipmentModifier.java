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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.io.xml.XMLReader;
import com.trollworks.gcs.io.xml.XMLWriter;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.WeightValue;

import java.io.IOException;

public class EquipmentModifier extends Modifier {
    private static final int                         CURRENT_VERSION        = 1;
    /** The root tag. */
    public static final  String                      TAG_MODIFIER           = "eqp_modifier";
    /** The root tag for containers. */
    public static final  String                      TAG_MODIFIER_CONTAINER = "eqp_modifier_container";
    /** The attribute for the cost. */
    public static final  String                      TAG_COST_ADJ           = "cost";
    /** The attribute for the cost type. */
    public static final  String                      ATTRIBUTE_COST_TYPE    = "cost_type";
    /** The attribute for the weight. */
    public static final  String                      TAG_WEIGHT_ADJ         = "weight";
    /** The attribute for the weight type. */
    public static final  String                      ATTRIBUTE_WEIGHT_TYPE  = "weight_type";
    /** The notification prefix used. */
    public static final  String                      PREFIX                 = GURPSCharacter.CHARACTER_PREFIX + "eqpmod" + Notifier.SEPARATOR;
    /** The notification ID for enabled changes. */
    public static final  String                      ID_ENABLED             = PREFIX + ATTRIBUTE_ENABLED;
    /** The field ID for when the categories change. */
    public static final  String                      ID_CATEGORY            = PREFIX + "Category";
    /** The notification ID for list changes. */
    public static final  String                      ID_LIST_CHANGED        = PREFIX + "list_changed";
    /** The notification ID for cost adjustment changes. */
    public static final  String                      ID_COST_ADJ            = PREFIX + TAG_COST_ADJ;
    /** The notification ID for weight adjustment changes. */
    public static final  String                      ID_WEIGHT_ADJ          = PREFIX + TAG_WEIGHT_ADJ;
    private              EquipmentModifierCostType   mCostType;
    private              Fixed6                      mCostAmount;
    private              EquipmentModifierWeightType mWeightType;
    private              Fixed6                      mWeightMultiplier;
    private              WeightValue                 mWeightAddition;

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
        mWeightMultiplier = other.mWeightMultiplier;
        mWeightAddition = new WeightValue(other.mWeightAddition);
        if (deep) {
            int count = other.getChildCount();
            for (int i = 0; i < count; i++) {
                addChild(new EquipmentModifier(file, (EquipmentModifier) other.getChild(i), true));
            }
        }
    }

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file   The {@link DataFile} to use.
     * @param reader The {@link XMLReader} to use.
     * @param state  The {@link LoadState} to use.
     */
    public EquipmentModifier(DataFile file, XMLReader reader, LoadState state) throws IOException {
        this(file, TAG_MODIFIER_CONTAINER.equals(reader.getName()));
        load(reader, state);
    }

    /**
     * Creates a new {@link EquipmentModifier}.
     *
     * @param file        The {@link DataFile} to use.
     * @param isContainer Whether or not this row allows children.
     */
    public EquipmentModifier(DataFile file, boolean isContainer) {
        super(file, isContainer);
        mCostType = EquipmentModifierCostType.COST_FACTOR;
        mCostAmount = Fixed6.ONE;
        mWeightType = EquipmentModifierWeightType.MULTIPLIER;
        mWeightMultiplier = Fixed6.ONE;
        mWeightAddition = new WeightValue(Fixed6.ZERO, DisplayPreferences.getWeightUnits());
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
    public String getNotificationPrefix() {
        return PREFIX;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof EquipmentModifier && super.isEquivalentTo(obj)) {
            EquipmentModifier row = (EquipmentModifier) obj;
            return mCostType == row.mCostType && mCostAmount.equals(row.mCostAmount) && mWeightType == row.mWeightType && mWeightMultiplier.equals(row.mWeightMultiplier) && mWeightAddition.equals(row.mWeightAddition);
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
    public Fixed6 getCostAdjAmount() {
        return mCostAmount;
    }

    /**
     * @param costAmount The amount for the cost modifier to set.
     * @return {@code true} if a change was made.
     */
    public boolean setCostAdjAmount(Fixed6 costAmount) {
        if (!mCostAmount.equals(costAmount)) {
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
    public Fixed6 getWeightAdjMultiplier() {
        return mWeightMultiplier;
    }

    /**
     * @param multiplier The amount for the weight multiplier. Only valid if {@code
     *                   getWeightAdjType() == EquipmentModifierWeightType.MULTIPLIER}.
     * @return {@code true} if a change was made.
     */
    public boolean setWeightAdjMultiplier(Fixed6 multiplier) {
        if (mWeightType == EquipmentModifierWeightType.MULTIPLIER && !mWeightMultiplier.equals(multiplier)) {
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
        return canHaveChildren() ? TAG_MODIFIER_CONTAINER : TAG_MODIFIER;
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
        mCostAmount = Fixed6.ONE;
        mWeightType = EquipmentModifierWeightType.MULTIPLIER;
        mWeightMultiplier = Fixed6.ONE;
        mWeightAddition = new WeightValue(Fixed6.ZERO, DisplayPreferences.getWeightUnits());
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (!state.mForUndo && (TAG_MODIFIER.equals(name) || TAG_MODIFIER_CONTAINER.equals(name))) {
            addChild(new EquipmentModifier(mDataFile, reader, state));
        } else if (!canHaveChildren()) {
            if (TAG_COST_ADJ.equals(name)) {
                mCostType = Enums.extract(reader.getAttribute(ATTRIBUTE_COST_TYPE), EquipmentModifierCostType.values(), EquipmentModifierCostType.COST_FACTOR);
                mCostAmount = new Fixed6(reader.readText(), Fixed6.ONE, false);
            } else if (TAG_WEIGHT_ADJ.equals(name)) {
                mWeightType = Enums.extract(reader.getAttribute(ATTRIBUTE_WEIGHT_TYPE), EquipmentModifierWeightType.values(), EquipmentModifierWeightType.MULTIPLIER);
                switch (mWeightType) {
                case BASE_ADDITION:
                case FINAL_ADDITION:
                    mWeightAddition = WeightValue.extract(reader.readText(), false);
                    break;
                case MULTIPLIER:
                default:
                    mWeightMultiplier = new Fixed6(reader.readText(), Fixed6.ONE, false);
                    break;
                }
            } else {
                super.loadSubElement(reader, state);
            }
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void saveSelf(XMLWriter out, boolean forUndo) {
        super.saveSelf(out, forUndo);
        if (!canHaveChildren()) {
            out.simpleTagWithAttribute(TAG_COST_ADJ, mCostAmount.toString(), ATTRIBUTE_COST_TYPE, Enums.toId(mCostType));
            switch (mWeightType) {
            case BASE_ADDITION:
            case FINAL_ADDITION:
                if (!mWeightAddition.getNormalizedValue().equals(Fixed6.ZERO)) {
                    out.simpleTagWithAttribute(TAG_WEIGHT_ADJ, mWeightAddition.toString(false), ATTRIBUTE_WEIGHT_TYPE, Enums.toId(mWeightType));
                }
                break;
            case MULTIPLIER:
            default:
                if (!mWeightMultiplier.equals(Fixed6.ONE)) {
                    out.simpleTagWithAttribute(TAG_WEIGHT_ADJ, mWeightMultiplier.toString(), ATTRIBUTE_WEIGHT_TYPE, Enums.toId(mWeightType));
                }
                break;
            }
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
        return canHaveChildren() ? "" : mCostType.format(mCostAmount) + " " + mCostType;
    }

    /** @return The formatted weight adjustment. */
    public String getWeightDescription() {
        if (canHaveChildren()) {
            return "";
        }
        StringBuilder builder = new StringBuilder();
        switch (mWeightType) {
        case BASE_ADDITION:
        case FINAL_ADDITION:
            if (!mWeightAddition.getNormalizedValue().equals(Fixed6.ZERO)) {
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
            if (!mWeightMultiplier.equals(Fixed6.ONE)) {
                builder.append(mWeightMultiplier.toLocalizedString());
                builder.append(' ');
                builder.append(mWeightType);
            }
            break;
        }
        return builder.toString();
    }

    @Override
    protected String getCategoryID() {
        return ID_CATEGORY;
    }
}
