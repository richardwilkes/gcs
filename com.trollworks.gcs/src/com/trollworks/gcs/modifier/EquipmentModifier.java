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
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.utility.xml.XMLReader;
import com.trollworks.gcs.utility.xml.XMLWriter;

import java.io.IOException;

public class EquipmentModifier extends Modifier {
    private static final int                         CURRENT_VERSION        = 2;
    /** The root tag. */
    public static final  String                      TAG_MODIFIER           = "eqp_modifier";
    /** The root tag for containers. */
    public static final  String                      TAG_MODIFIER_CONTAINER = "eqp_modifier_container";
    /** The attribute for the cost. */
    public static final  String                      TAG_COST_ADJ           = "cost";
    private static final String                      TAG_TECH_LEVEL         = "tech_level";
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
    /** The field ID for tech level changes. */
    public static final  String                      ID_TECH_LEVEL          = PREFIX + "TechLevel";
    private static final String                      DEFAULT_COST_AMOUNT    = "+0";
    private              EquipmentModifierCostType   mCostType;
    private              String                      mCostAmount;
    private              EquipmentModifierWeightType mWeightType;
    private              String                      mWeightAmount;
    private              String                      mTechLevel;

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
        mCostType = EquipmentModifierCostType.TO_ORIGINAL_COST;
        mCostAmount = DEFAULT_COST_AMOUNT;
        mWeightType = EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT;
        mWeightAmount = getDefaultWeightAmount();
        mTechLevel = "";
    }

    private String getDefaultWeightAmount() {
        return "+" + new WeightValue(Fixed6.ZERO, getDataFile().defaultWeightUnits()).toString();
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
            notifySingle(ID_COST_ADJ);
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
        mCostType = EquipmentModifierCostType.TO_ORIGINAL_COST;
        mCostAmount = DEFAULT_COST_AMOUNT;
        mWeightType = EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT;
        mWeightAmount = getDefaultWeightAmount();
        mTechLevel = "";
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (!state.mForUndo && (TAG_MODIFIER.equals(name) || TAG_MODIFIER_CONTAINER.equals(name))) {
            addChild(new EquipmentModifier(mDataFile, reader, state));
        } else if (!canHaveChildren()) {
            if (TAG_COST_ADJ.equals(name)) {
                mCostType = Enums.extract(reader.getAttribute(ATTRIBUTE_COST_TYPE), EquipmentModifierCostType.values(), EquipmentModifierCostType.TO_ORIGINAL_COST);
                mCostAmount = mCostType.format(reader.readText(), false);
            } else if (TAG_WEIGHT_ADJ.equals(name)) {
                mWeightType = Enums.extract(reader.getAttribute(ATTRIBUTE_WEIGHT_TYPE), EquipmentModifierWeightType.values(), EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT);
                mWeightAmount = mWeightType.format(reader.readText(), getDataFile().defaultWeightUnits(), false);
            } else if (TAG_TECH_LEVEL.equals(name)) {
                mTechLevel = reader.readText().replace("\n", " ");
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
            if (mCostType != EquipmentModifierCostType.TO_ORIGINAL_COST || !mCostAmount.equals(DEFAULT_COST_AMOUNT)) {
                out.simpleTagWithAttribute(TAG_COST_ADJ, mCostAmount, ATTRIBUTE_COST_TYPE, Enums.toId(mCostType));
            }
            if (mWeightType != EquipmentModifierWeightType.TO_ORIGINAL_WEIGHT || !mWeightAmount.equals(getDefaultWeightAmount())) {
                out.simpleTagWithAttribute(TAG_WEIGHT_ADJ, mWeightAmount, ATTRIBUTE_WEIGHT_TYPE, Enums.toId(mWeightType));
            }
            out.simpleTagNotEmpty(TAG_TECH_LEVEL, mTechLevel);
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

    @Override
    protected String getCategoryID() {
        return ID_CATEGORY;
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
}
