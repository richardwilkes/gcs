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
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** Model for trait modifiers */
public class AdvantageModifier extends Modifier {
    private static final int                       CURRENT_JSON_VERSION   = 1;
    private static final int                       CURRENT_VERSION        = 2;
    /** The root tag. */
    public static final  String                    TAG_MODIFIER           = "modifier";
    /** The root tag for containers. */
    public static final  String                    TAG_MODIFIER_CONTAINER = "modifier_container";
    /** The tag for the base cost. */
    public static final  String                    TAG_COST               = "cost";
    /** The attribute for the cost type. */
    public static final  String                    ATTRIBUTE_COST_TYPE    = "type";
    private static final  String                    KEY_COST_TYPE    = "cost_type";
    /** The tag for the cost per level. */
    public static final  String                    TAG_LEVELS             = "levels";
    /** The tag for how the cost is affected. */
    public static final  String                    TAG_AFFECTS            = "affects";
    /** The notification prefix used. */
    public static final  String                    PREFIX                 = GURPSCharacter.CHARACTER_PREFIX + "advmod" + Notifier.SEPARATOR;
    /** The field ID for when the categories change. */
    public static final  String                    ID_CATEGORY            = PREFIX + "Category";
    /** The field ID for enabled changes. */
    public static final  String                    ID_ENABLED             = PREFIX + ATTRIBUTE_ENABLED;
    /** The field ID for list changes. */
    public static final  String                    ID_LIST_CHANGED        = PREFIX + "list_changed";
    /** The cost type of the {@link AdvantageModifier}. */
    protected            AdvantageModifierCostType mCostType;
    private              int                       mCost;
    private              double                    mCostMultiplier;
    private              int                       mLevels;
    private              Affects                   mAffects;

    /**
     * Creates a new {@link AdvantageModifier}.
     *
     * @param file  The {@link DataFile} to use.
     * @param other Another {@link AdvantageModifier} to clone.
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

    public AdvantageModifier(DataFile file, JsonMap m, LoadState state) throws IOException {
        this(file, TAG_MODIFIER_CONTAINER.equals(m.getString(DataFile.KEY_TYPE)));
        load(m, state);
    }

    /**
     * Creates a new {@link AdvantageModifier}.
     *
     * @param file   The {@link DataFile} to use.
     * @param reader The {@link XMLReader} to use.
     * @param state  The {@link LoadState} to use.
     */
    public AdvantageModifier(DataFile file, XMLReader reader, LoadState state) throws IOException {
        this(file, TAG_MODIFIER_CONTAINER.equals(reader.getName()));
        load(reader, state);
    }

    /**
     * Creates a new {@link AdvantageModifier}.
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
    public String getNotificationPrefix() {
        return PREFIX;
    }

    @Override
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof AdvantageModifier && super.isEquivalentTo(obj)) {
            AdvantageModifier row = (AdvantageModifier) obj;
            return mLevels == row.mLevels && mCost == row.mCost && mCostMultiplier == row.mCostMultiplier && mCostType == row.mCostType && mAffects == row.mAffects;
        }
        return false;
    }

    /** @return An exact clone of this modifier. */
    public AdvantageModifier cloneModifier(boolean deep) {
        return new AdvantageModifier(mDataFile, this, deep);
    }

    @Override
    public RetinaIcon getIcon(boolean marker) {
        return marker ? Images.ADM_MARKER : Images.ADM_FILE;
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
            notifySingle(getNotificationPrefix() + TAG_COST);
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
            notifySingle(getNotificationPrefix() + TAG_COST);
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
            notifySingle(getNotificationPrefix() + TAG_COST);
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
            notifySingle(getNotificationPrefix() + TAG_COST);
            return true;
        }
        return false;
    }

    /** @return {@code true} if this {@link AdvantageModifier} has levels. */
    public boolean hasLevels() {
        return mCostType == AdvantageModifierCostType.PERCENTAGE && mLevels > 0;
    }

    @Override
    public RowEditor<AdvantageModifier> createEditor() {
        return new AdvantageModifierEditor(this);
    }

    @Override
    public String getJSONTypeName() {
        return canHaveChildren() ? TAG_MODIFIER_CONTAINER : TAG_MODIFIER;
    }

    @Override
    public int getJSONVersion() {
        return CURRENT_JSON_VERSION;
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
    protected void prepareForLoad(LoadState state) {
        super.prepareForLoad(state);
        mCostType = AdvantageModifierCostType.PERCENTAGE;
        mCost = 0;
        mCostMultiplier = 1.0;
        mLevels = 0;
        mAffects = Affects.TOTAL;
    }

    @Override
    protected void loadSubElement(XMLReader reader, LoadState state) throws IOException {
        String name = reader.getName();
        if (!state.mForUndo && (TAG_MODIFIER.equals(name) || TAG_MODIFIER_CONTAINER.equals(name))) {
            addChild(new AdvantageModifier(mDataFile, reader, state));
        } else if (!canHaveChildren()) {
            if (TAG_COST.equals(name)) {
                mCostType = Enums.extract(reader.getAttribute(ATTRIBUTE_COST_TYPE), AdvantageModifierCostType.values(), AdvantageModifierCostType.PERCENTAGE);
                if (mCostType == AdvantageModifierCostType.MULTIPLIER) {
                    mCostMultiplier = reader.readDouble(1.0);
                } else {
                    mCost = reader.readInteger(0);
                }
            } else if (TAG_LEVELS.equals(name)) {
                mLevels = reader.readInteger(0);
            } else if (TAG_AFFECTS.equals(name)) {
                mAffects = Enums.extract(reader.readText(), Affects.values(), Affects.TOTAL);
            } else {
                super.loadSubElement(reader, state);
            }
        } else {
            super.loadSubElement(reader, state);
        }
    }

    @Override
    protected void loadSelf(JsonMap m, LoadState state) throws IOException {
        super.loadSelf(m, state);
        if (!canHaveChildren()) {
            mCostType = Enums.extract(m.getString(KEY_COST_TYPE), AdvantageModifierCostType.values(), AdvantageModifierCostType.PERCENTAGE);
            if (mCostType == AdvantageModifierCostType.MULTIPLIER) {
                mCostMultiplier = m.getDouble(TAG_COST);
            } else {
                mCost = m.getInt(TAG_COST);
                mAffects = Enums.extract(m.getString(TAG_AFFECTS), Affects.values(), Affects.TOTAL);
            }
            mLevels = m.getInt(TAG_LEVELS);
        }
    }

    @Override
    protected void loadChild(JsonMap m, LoadState state) throws IOException {
        if (!state.mForUndo) {
            String type = m.getString(DataFile.KEY_TYPE);
            if (TAG_MODIFIER.equals(type) || TAG_MODIFIER_CONTAINER.equals(type)) {
                addChild(new AdvantageModifier(mDataFile, m, state));
            } else {
                Log.warn("invalid child type: " + type);
            }
        }
    }

    @Override
    protected void saveSelf(JsonWriter w, boolean forUndo) throws IOException {
        super.saveSelf(w, forUndo);
        if (!canHaveChildren()) {
            w.keyValue(KEY_COST_TYPE, Enums.toId(mCostType));
            if (mCostType == AdvantageModifierCostType.MULTIPLIER) {
                w.keyValue(TAG_COST, mCostMultiplier);
            } else {
                w.keyValue(TAG_COST, mCost);
                w.keyValue(TAG_AFFECTS, Enums.toId(mAffects));
            }
            w.keyValueNot(TAG_LEVELS, mLevels, 0);
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

    /** @return A full description of this {@link AdvantageModifier}. */
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

    /** @return The formatted cost. */
    public String getCostDescription() {
        StringBuilder             builder  = new StringBuilder();
        AdvantageModifierCostType costType = getCostType();
        switch (costType) {
        case PERCENTAGE:
        case POINTS:
        default:
            builder.append(Numbers.formatWithForcedSign(getCostModifier()));
            if (costType == AdvantageModifierCostType.PERCENTAGE) {
                builder.append('%');
            }
            String desc = mAffects.getShortTitle();
            if (!desc.isEmpty()) {
                builder.append(' ');
                builder.append(desc);
            }
            break;
        case MULTIPLIER:
            builder.append('x');
            builder.append(Numbers.format(getCostMultiplier()));
            break;
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
            notifySingle(getNotificationPrefix() + TAG_AFFECTS);
            return true;
        }
        return false;
    }

    @Override
    protected String getCategoryID() {
        return ID_CATEGORY;
    }
}
