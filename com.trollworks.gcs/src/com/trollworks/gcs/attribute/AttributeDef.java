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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.character.CharacterVariableResolver;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AttributeDef implements Cloneable, Comparable<AttributeDef> {
    private static final String KEY_ID                      = "id";
    private static final String KEY_NAME                    = "name";
    private static final String KEY_DESCRIPTION             = "description";
    private static final String KEY_ATTRIBUTE_BASE          = "attribute_base";
    private static final String KEY_COST_PER_POINT          = "cost_per_point";
    private static final String KEY_COST_ADJ_PERCENT_PER_SM = "cost_adj_percent_per_sm";
    private static final String KEY_DECIMAL                 = "decimal";

    private String  mID;
    private String  mName;
    private String  mDescription;
    private String  mAttributeBase;
    private int     mOrder;
    private int     mCostPerPoint;
    private int     mCostAdjPercentPerSM;
    private boolean mDecimal;

    public static final Map<String, AttributeDef> cloneMap(Map<String, AttributeDef> m) {
        Map<String, AttributeDef> result = new HashMap<>();
        for (AttributeDef def : m.values()) {
            result.put(def.getID(), def.clone());
        }
        return result;
    }

    public static final Map<String, AttributeDef> load(JsonArray a) {
        Map<String, AttributeDef> m      = new HashMap<>();
        int                       length = a.size();
        for (int i = 0; i < length; i++) {
            AttributeDef def = new AttributeDef(a.getMap(i), i);
            m.put(def.getID(), def);
        }
        return m;
    }

    public static final List<AttributeDef> getOrdered(Map<String, AttributeDef> m) {
        List<AttributeDef> ordered = new ArrayList<>(m.values());
        Collections.sort(ordered);
        return ordered;
    }

    public static final void writeOrdered(JsonWriter w, Map<String, AttributeDef> m) throws IOException {
        w.startArray();
        for (AttributeDef def : getOrdered(m)) {
            def.toJSON(w);
        }
        w.endArray();

    }

    public static final Map<String, AttributeDef> createStandardAttributes() {
        Map<String, AttributeDef> m = new HashMap<>();
        int                       i = 0;
        m.put("st", new AttributeDef("st", I18n.Text("ST"), I18n.Text("Strength"), "10", ++i, 10, 10, false));
        m.put("dx", new AttributeDef("dx", I18n.Text("DX"), I18n.Text("Dexterity"), "10", ++i, 20, 0, false));
        m.put("iq", new AttributeDef("iq", I18n.Text("IQ"), I18n.Text("Intelligence"), "10", ++i, 20, 0, false));
        m.put("ht", new AttributeDef("ht", I18n.Text("HT"), I18n.Text("Health"), "10", ++i, 10, 0, false));
        m.put("will", new AttributeDef("will", I18n.Text("Will"), I18n.Text("Will"), "$" + Attribute.ID_ATTR_PREFIX + "iq", ++i, 5, 0, false));
        m.put("fright_check", new AttributeDef("fright_check", I18n.Text("Fright Check"), I18n.Text("Fright Check"), "$" + Attribute.ID_ATTR_PREFIX + "will", ++i, 2, 0, false));
        m.put("per", new AttributeDef("per", I18n.Text("Per"), I18n.Text("Perception"), "$" + Attribute.ID_ATTR_PREFIX + "iq", ++i, 5, 0, false));
        m.put("vision", new AttributeDef("vision", I18n.Text("Vision"), I18n.Text("Vision"), "$" + Attribute.ID_ATTR_PREFIX + "per", ++i, 2, 0, false));
        m.put("hearing", new AttributeDef("hearing", I18n.Text("Hearing"), I18n.Text("Hearing"), "$" + Attribute.ID_ATTR_PREFIX + "per", ++i, 2, 0, false));
        m.put("taste_smell", new AttributeDef("taste_smell", I18n.Text("Taste & Smell"), I18n.Text("Taste & Smell"), "$" + Attribute.ID_ATTR_PREFIX + "per", ++i, 2, 0, false));
        m.put("touch", new AttributeDef("touch", I18n.Text("Touch"), I18n.Text("Touch"), "$" + Attribute.ID_ATTR_PREFIX + "per", ++i, 2, 0, false));
        m.put("basic_speed", new AttributeDef("basic_speed", I18n.Text("Basic Speed"), I18n.Text("Basic Speed"), "($" + Attribute.ID_ATTR_PREFIX + "dx+$" + Attribute.ID_ATTR_PREFIX + "ht)/4", ++i, 20, 0, true));
        m.put("basic_move", new AttributeDef("basic_move", I18n.Text("Basic Move"), I18n.Text("Basic Move"), "floor($" + Attribute.ID_ATTR_PREFIX + "basic_speed)", ++i, 5, 0, false));
        return m;
    }

    public AttributeDef(String id, String name, String desc, String attributeBase, int order, int costPerPoint, int costAdjPercentPerSM, boolean decimal) {
        mID = id;
        mName = name;
        mDescription = desc;
        mAttributeBase = attributeBase;
        mOrder = order;
        mCostPerPoint = costPerPoint;
        mCostAdjPercentPerSM = costAdjPercentPerSM;
        mDecimal = decimal;
    }

    public AttributeDef(JsonMap m, int order) {
        mID = m.getString(KEY_ID);
        mName = m.getString(KEY_NAME);
        mDescription = m.getString(KEY_DESCRIPTION);
        mAttributeBase = m.getString(KEY_ATTRIBUTE_BASE);
        mOrder = order;
        mCostPerPoint = m.getInt(KEY_COST_PER_POINT);
        mCostAdjPercentPerSM = m.getInt(KEY_COST_ADJ_PERCENT_PER_SM);
        mDecimal = m.getBoolean(KEY_DECIMAL);
    }

    public String getID() {
        return mID;
    }

    public void setID(String ID) {
        mID = ID;
    }

    public String getName() {
        return mName;
    }

    public void setName(String name) {
        mName = name;
    }

    public String getDescription() {
        return mDescription;
    }

    public void setDescription(String description) {
        mDescription = description;
    }

    public String getAttributeBase() {
        return mAttributeBase;
    }

    public void setAttributeBase(String attributeBase) {
        mAttributeBase = attributeBase;
    }

    public int getOrder() {
        return mOrder;
    }

    public void setOrder(int order) {
        mOrder = order;
    }

    public int getCostPerPoint() {
        return mCostPerPoint;
    }

    public void setCostPerPoint(int costPerPoint) {
        mCostPerPoint = costPerPoint;
    }

    public int getCostAdjPercentPerSM() {
        return mCostAdjPercentPerSM;
    }

    public void setCostAdjPercentPerSM(int costAdjPercentPerSM) {
        mCostAdjPercentPerSM = costAdjPercentPerSM;
    }

    public boolean isDecimal() {
        return mDecimal;
    }

    public boolean isPrimary() {
        try {
            Integer.parseInt(mAttributeBase);
            return true;
        } catch (NumberFormatException ex) {
            return false;
        }
    }

    public double getBaseValue(CharacterVariableResolver resolver) {
        Evaluator evaluator = new Evaluator(resolver);
        String    exclude   = Attribute.ID_ATTR_PREFIX + mID;
        resolver.addExclusion(exclude);
        double value;
        try {
            value = evaluator.evaluateToNumber(mAttributeBase);
        } catch (EvaluationException ex) {
            Log.error(ex);
            value = 0;
        }
        resolver.removeExclusion(exclude);
        return value;
    }

    public int computeCost(GURPSCharacter character, double value, int sm, int costReduction) {
        if (sm > 0 && mCostAdjPercentPerSM > 0) {
            costReduction += sm * mCostAdjPercentPerSM;
            if (costReduction < 0) {
                costReduction = 0;
            } else if (costReduction < 80) {
                costReduction = 80;
            }
        }
        int cost = (int) Math.floor(mCostPerPoint * value);
        return costReduction == 0 ? cost : (99 + cost * (100 - costReduction)) / 100;
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_ID, mID);
        w.keyValue(KEY_NAME, mName);
        w.keyValue(KEY_DESCRIPTION, mDescription);
        w.keyValue(KEY_ATTRIBUTE_BASE, mAttributeBase);
        w.keyValue(KEY_COST_PER_POINT, mCostPerPoint);
        w.keyValue(KEY_COST_ADJ_PERCENT_PER_SM, mCostAdjPercentPerSM);
        w.endMap();
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (!(o instanceof AttributeDef)) {
            return false;
        }
        AttributeDef that = (AttributeDef) o;
        if (mCostPerPoint != that.mCostPerPoint) {
            return false;
        }
        if (mCostAdjPercentPerSM != that.mCostAdjPercentPerSM) {
            return false;
        }
        if (!mID.equals(that.mID)) {
            return false;
        }
        if (!mName.equals(that.mName)) {
            return false;
        }
        if (!mDescription.equals(that.mDescription)) {
            return false;
        }
        return mAttributeBase.equals(that.mAttributeBase);
    }

    @Override
    public int hashCode() {
        int result = mID.hashCode();
        result = 31 * result + mName.hashCode();
        result = 31 * result + mDescription.hashCode();
        result = 31 * result + mAttributeBase.hashCode();
        result = 31 * result + mCostPerPoint;
        result = 31 * result + mCostAdjPercentPerSM;
        return result;
    }

    @Override
    protected AttributeDef clone() {
        try {
            return (AttributeDef) super.clone();
        } catch (CloneNotSupportedException e) {
            // This can't happen
            Log.error(e);
            System.exit(1);
        }
        return null;
    }

    @Override
    // Used solely for remembering the desired ordering.
    public int compareTo(AttributeDef other) {
        return Integer.compare(mOrder, other.mOrder);
    }
}
