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

package com.trollworks.gcs.character.attribute;

import com.trollworks.gcs.character.GURPSCharacter;
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

    private String mID;
    private String mName;
    private String mDescription;
    private String mAttributeBase;
    private int    mOrder;
    private int    mCostPerPoint;
    private int    mCostAdjPercentPerSM;

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
        m.put("st", new AttributeDef("st", I18n.Text("ST"), I18n.Text("Strength"), "10", 0, 10, 10));
        m.put("dx", new AttributeDef("dx", I18n.Text("DX"), I18n.Text("Dexterity"), "10", 1, 20, 0));
        m.put("iq", new AttributeDef("iq", I18n.Text("IQ"), I18n.Text("Intelligence"), "10", 2, 20, 0));
        m.put("ht", new AttributeDef("ht", I18n.Text("HT"), I18n.Text("Health"), "10", 3, 10, 0));
        return m;
    }

    public AttributeDef(String id, String name, String desc, String attributeBase, int order, int costPerPoint, int costAdjPercentPerSM) {
        mID = id;
        mName = name;
        mDescription = desc;
        mAttributeBase = attributeBase;
        mOrder = order;
        mCostPerPoint = costPerPoint;
        mCostAdjPercentPerSM = costAdjPercentPerSM;
    }

    public AttributeDef(JsonMap m, int order) {
        mID = m.getString(KEY_ID);
        mName = m.getString(KEY_NAME);
        mDescription = m.getString(KEY_DESCRIPTION);
        mAttributeBase = m.getString(KEY_ATTRIBUTE_BASE);
        mOrder = order;
        mCostPerPoint = m.getInt(KEY_COST_PER_POINT);
        mCostAdjPercentPerSM = m.getInt(KEY_COST_ADJ_PERCENT_PER_SM);
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

    public int getBaseValue(GURPSCharacter character) {
        if (mName.equals(mAttributeBase)) {
            return 0; // Referring to self, which isn't valid.
        }
        try {
            return Integer.parseInt(mAttributeBase);
        } catch (NumberFormatException ex) {
            Attribute attr = character.getAttributes().get(mAttributeBase);
            if (attr == null) {
                return 0;
            }
            return attr.getValue(character);
        }
    }

    public int computeCost(GURPSCharacter character, int value, int sm, int costReduction) {
        if (value < 0) {
            return 0;
        }
        int cost = mCostPerPoint * value;
        if (sm > 0 && mCostAdjPercentPerSM > 0) {
            costReduction += sm * mCostAdjPercentPerSM;
            if (costReduction < 0) {
                costReduction = 0;
            } else if (costReduction < 80) {
                costReduction = 80;
            }
        }
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
