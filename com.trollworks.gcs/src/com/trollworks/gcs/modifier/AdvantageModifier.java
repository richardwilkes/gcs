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

import com.trollworks.gcs.character.CollectedModels;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;

import java.io.IOException;
import javax.swing.Icon;

/** Model for trait modifiers */
public class AdvantageModifier extends Modifier {
    public static final  String KEY_MODIFIER           = "modifier";
    public static final  String KEY_MODIFIER_CONTAINER = "modifier_container";
    public static final  String KEY_COST               = "cost";
    private static final String KEY_COST_TYPE          = "cost_type";
    public static final  String KEY_LEVELS             = "levels";
    public static final  String KEY_AFFECTS            = "affects";

    protected AdvantageModifierCostType mCostType;
    private   int                       mCost;
    private   double                    mCostMultiplier;
    private   int                       mLevels;
    private   Affects                   mAffects;

    /**
     * Creates a new AdvantageModifier.
     *
     * @param file  The {@link DataFile} to use.
     * @param other Another AdvantageModifier to clone.
     * @param deep  Whether or not to clone the children, grandchildren, etc.
     */
    public AdvantageModifier(DataFile file, AdvantageModifier other, boolean deep) {
        super(file, other);
        mCostType = other.mCostType;
        mCost = other.mCost;
        mCostMultiplier = other.mCostMultiplier;
        mLevels = other.mLevels;
        mAffects = other.mAffects;
        if (deep) {
            int count = other.getChildCount();
            for (int i = 0; i < count; i++) {
                addChild(new AdvantageModifier(file, (AdvantageModifier) other.getChild(i), true));
            }
        }
    }

    public AdvantageModifier(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        this(dataFile, KEY_MODIFIER_CONTAINER.equals(m.getString(DataFile.TYPE)));
        load(dataFile, m, state);
    }

    /**
     * Creates a new AdvantageModifier.
     *
     * @param file        The {@link DataFile} to use.
     * @param isContainer Whether or not this row allows children.
     */
    public AdvantageModifier(DataFile file, boolean isContainer) {
        super(file, isContainer);
        mCostType = AdvantageModifierCostType.PERCENTAGE;
        mCost = 0;
        mCostMultiplier = 1.0;
        mLevels = 0;
        mAffects = Affects.TOTAL;
    }

    @Override
    public AdvantageModifier cloneRow(DataFile newOwner, boolean deep, boolean forSheet) {
        return new AdvantageModifier(newOwner, this, deep);
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof AdvantageModifier row && super.isEquivalentTo(obj)) {
            return mLevels == row.mLevels && mCost == row.mCost && mCostMultiplier == row.mCostMultiplier && mCostType == row.mCostType && mAffects == row.mAffects;
        }
        return false;
    }

    /** @return An exact clone of this modifier. */
    public AdvantageModifier cloneModifier(boolean deep) {
        return new AdvantageModifier(mDataFile, this, deep);
    }

    @Override
    public Icon getIcon() {
        return FileType.ADVANTAGE_MODIFIER.getIcon();
    }

    /** @return The total cost modifier. */
    public int getCostModifier() {
        return mLevels > 0 ? mCost * mLevels : mCost;
    }

    /** @return The costType. */
    public AdvantageModifierCostType getCostType() {
        return mCostType;
    }

    /**
     * @param costType The value to set for costType.
     * @return Whether it was changed.
     */
    public boolean setCostType(AdvantageModifierCostType costType) {
        if (costType != mCostType) {
            mCostType = costType;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The cost. */
    public int getCost() {
        return mCost;
    }

    /**
     * @param cost The value to set for cost modifier.
     * @return Whether it was changed.
     */
    public boolean setCost(int cost) {
        if (mCost != cost) {
            mCost = cost;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The total cost multiplier. */
    public double getCostMultiplier() {
        return mCostMultiplier;
    }

    /**
     * @param multiplier The value to set for the cost multiplier.
     * @return Whether it was changed.
     */
    public boolean setCostMultiplier(double multiplier) {
        if (mCostMultiplier != multiplier) {
            mCostMultiplier = multiplier;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return The levels. */
    public int getLevels() {
        return mLevels;
    }

    /**
     * @param levels The value to set for cost modifier.
     * @return Whether it was changed.
     */
    public boolean setLevels(int levels) {
        if (levels < 0) {
            levels = 0;
        }
        if (mLevels != levels) {
            mLevels = levels;
            notifyOfChange();
            return true;
        }
        return false;
    }

    /** @return {@code true} if this AdvantageModifier has levels. */
    public boolean hasLevels() {
        return mCostType == AdvantageModifierCostType.PERCENTAGE && mLevels > 0;
    }

    @Override
    public RowEditor<AdvantageModifier> createEditor() {
        return new AdvantageModifierEditor(this);
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? KEY_MODIFIER_CONTAINER : KEY_MODIFIER;
    }

    @Override
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mCostType = AdvantageModifierCostType.PERCENTAGE;
        mCost = 0;
        mCostMultiplier = 1.0;
        mLevels = 0;
        mAffects = Affects.TOTAL;
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        if (!canHaveChildren()) {
            mCostType = Enums.extract(m.getString(KEY_COST_TYPE), AdvantageModifierCostType.values(), AdvantageModifierCostType.PERCENTAGE);
            if (mCostType == AdvantageModifierCostType.MULTIPLIER) {
                mCostMultiplier = m.getDouble(KEY_COST);
            } else {
                mCost = m.getInt(KEY_COST);
                mAffects = Enums.extract(m.getString(KEY_AFFECTS), Affects.values(), Affects.TOTAL);
            }
            mLevels = m.getInt(KEY_LEVELS);
        }
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.TYPE);
            if (KEY_MODIFIER.equals(type) || KEY_MODIFIER_CONTAINER.equals(type)) {
                addChild(new AdvantageModifier(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, SaveType saveType) throws IOException {
        super.saveSelf(w, saveType);
        if (!canHaveChildren()) {
            w.keyValue(KEY_COST_TYPE, Enums.toId(mCostType));
            if (mCostType == AdvantageModifierCostType.MULTIPLIER) {
                w.keyValue(KEY_COST, mCostMultiplier);
            } else {
                w.keyValue(KEY_COST, mCost);
                w.keyValue(KEY_AFFECTS, Enums.toId(mAffects));
            }
            w.keyValueNot(KEY_LEVELS, mLevels, 0);
        }
    }

    @Override
    public Object getData(Column column) {
        return AdvantageModifierColumnID.values()[column.getID()].getData(this);
    }

    @Override
    public String getDataAsText(Column column) {
        return AdvantageModifierColumnID.values()[column.getID()].getDataAsText(this);
    }

    @Override
    public String toString() {
        StringBuilder builder = new StringBuilder();
        builder.append(super.toString());
        if (hasLevels()) {
            builder.append(' ');
            builder.append(getLevels());
        }
        return builder.toString();
    }

    /** @return A full description of this AdvantageModifier. */
    public String getFullDescription() {
        StringBuilder builder = new StringBuilder();
        String        modNote = getNotes();
        builder.append(this);
        if (!modNote.isEmpty()) {
            builder.append(" (");
            builder.append(modNote);
            builder.append(')');
        }
        if ((mDataFile instanceof CollectedModels) && mDataFile.getSheetSettings().showAdvantageModifierAdj()) {
            builder.append(" [");
            builder.append(getCostDescription());
            builder.append(']');
        }
        return builder.toString();
    }

    /** @return The formatted cost. */
    public String getCostDescription() {
        StringBuilder             builder  = new StringBuilder();
        AdvantageModifierCostType costType = getCostType();
        if (costType == AdvantageModifierCostType.MULTIPLIER) {
            builder.append('x');
            builder.append(Numbers.format(getCostMultiplier()));
        } else {
            builder.append(Numbers.formatWithForcedSign(getCostModifier()));
            if (costType == AdvantageModifierCostType.PERCENTAGE) {
                builder.append('%');
            }
            String desc = mAffects.getShortTitle();
            if (!desc.isEmpty()) {
                builder.append(' ');
                builder.append(desc);
            }
        }
        return builder.toString();
    }

    /** @return The {@link Affects} setting. */
    public Affects getAffects() {
        return mAffects;
    }

    /**
     * @param affects The new {@link Affects} setting.
     * @return {@code true} if the setting changed.
     */
    public boolean setAffects(Affects affects) {
        if (affects != mAffects) {
            mAffects = affects;
            notifyOfChange();
            return true;
        }
        return false;
    }

    @Override
    public String getToolTip(Column column) {
        return AdvantageModifierColumnID.values()[column.getID()].getToolTip(this);
    }
}
