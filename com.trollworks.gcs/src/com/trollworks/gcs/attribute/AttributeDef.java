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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;
import com.trollworks.gcs.expression.VariableResolver;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.ID;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;

public class AttributeDef implements Cloneable, Comparable<AttributeDef> {
    private static final String KEY_ID                      = "id";
    private static final String KEY_TYPE                    = "type";
    private static final String KEY_SHORT_NAME              = "name";
    private static final String KEY_FULL_NAME               = "full_name";
    private static final String KEY_ATTRIBUTE_BASE          = "attribute_base";
    private static final String KEY_COST_PER_POINT          = "cost_per_point";
    private static final String KEY_COST_ADJ_PERCENT_PER_SM = "cost_adj_percent_per_sm";
    private static final String KEY_THRESHOLDS              = "thresholds";

    public static final Set<String> RESERVED = new HashSet<>();

    private String              mID;
    private AttributeType       mType;
    private String              mName;
    private String              mFullName;
    private String              mAttributeBase;
    private int                 mOrder;
    private int                 mCostPerPoint;
    private int                 mCostAdjPercentPerSM;
    private List<PoolThreshold> mThresholds;

    static {
        RESERVED.add("skill");
        RESERVED.add("parry");
        RESERVED.add("block");
        RESERVED.add("dodge");
        RESERVED.add("sm");
    }

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

        // Primary attributes
        m.put("st", new AttributeDef("st", AttributeType.INTEGER, I18n.text("ST"), I18n.text("Strength"), "10", ++i, 10, 10));
        m.put("dx", new AttributeDef("dx", AttributeType.INTEGER, I18n.text("DX"), I18n.text("Dexterity"), "10", ++i, 20, 0));
        m.put("iq", new AttributeDef("iq", AttributeType.INTEGER, I18n.text("IQ"), I18n.text("Intelligence"), "10", ++i, 20, 0));
        m.put("ht", new AttributeDef("ht", AttributeType.INTEGER, I18n.text("HT"), I18n.text("Health"), "10", ++i, 10, 0));

        // Secondary attributes
        m.put("will", new AttributeDef("will", AttributeType.INTEGER, I18n.text("Will"), "", "$iq", ++i, 5, 0));
        m.put("fright_check", new AttributeDef("fright_check", AttributeType.INTEGER, I18n.text("Fright Check"), "", "$will", ++i, 2, 0));
        m.put("per", new AttributeDef("per", AttributeType.INTEGER, I18n.text("Per"), I18n.text("Perception"), "$iq", ++i, 5, 0));
        m.put("vision", new AttributeDef("vision", AttributeType.INTEGER, I18n.text("Vision"), "", "$per", ++i, 2, 0));
        m.put("hearing", new AttributeDef("hearing", AttributeType.INTEGER, I18n.text("Hearing"), "", "$per", ++i, 2, 0));
        m.put("taste_smell", new AttributeDef("taste_smell", AttributeType.INTEGER, I18n.text("Taste & Smell"), "", "$per", ++i, 2, 0));
        m.put("touch", new AttributeDef("touch", AttributeType.INTEGER, I18n.text("Touch"), "", "$per", ++i, 2, 0));
        m.put("basic_speed", new AttributeDef("basic_speed", AttributeType.DECIMAL, I18n.text("Basic Speed"), "", "($dx+$ht)/4", ++i, 20, 0));
        m.put("basic_move", new AttributeDef("basic_move", AttributeType.INTEGER, I18n.text("Basic Move"), "", "floor($basic_speed)", ++i, 5, 0));

        // Point pools
        List<ThresholdOps> ops = new ArrayList<>();
        ops.add(ThresholdOps.HALVE_MOVE);
        ops.add(ThresholdOps.HALVE_DODGE);
        ops.add(ThresholdOps.HALVE_ST);

        List<PoolThreshold> thresholds = new ArrayList<>();
        thresholds.add(new PoolThreshold(-1, 1, 0, I18n.text("Unconscious"), "", ops));
        thresholds.add(new PoolThreshold(0, 1, 0, I18n.text("Collapse"), I18n.text("""
                <html><body>
                <b>Roll vs. Will</b> to do anything besides talk or rest; failure causes unconsciousness<br>
                Each FP you lose below 0 also causes 1 HP of injury<br>
                Move, Dodge and ST are halved (B426)
                </body></html>"""), ops));
        thresholds.add(new PoolThreshold(1, 3, 0, I18n.text("Tired"), I18n.text("Move, Dodge and ST are halved (B426)"), ops));
        thresholds.add(new PoolThreshold(1, 1, -1, I18n.text("Tiring"), "", null));
        thresholds.add(new PoolThreshold(1, 1, 0, I18n.text("Rested"), "", null));

        m.put("fp", new AttributeDef("fp", I18n.text("FP"), I18n.text("Fatigue Points"), "$ht", ++i, 3, 0, thresholds));

        ops = new ArrayList<>();
        ops.add(ThresholdOps.HALVE_MOVE);
        ops.add(ThresholdOps.HALVE_DODGE);

        thresholds = new ArrayList<>();
        thresholds.add(new PoolThreshold(-5, 1, 0, I18n.text("Dead"), "", ops));
        for (int j = -4; j < 0; j++) {
            thresholds.add(new PoolThreshold(j, 1, 0, String.format(I18n.text("Dying #%d"), Integer.valueOf(-j)), String.format(I18n.text("""
                    <html><body>
                    <b>Roll vs. HT</b> to avoid death<br>
                    <b>Roll vs. HT%d</b> every second to avoid falling unconscious<br>
                    Move and Dodge are halved (B419)
                    </body></html>"""), Integer.valueOf(j)), ops));
        }
        thresholds.add(new PoolThreshold(0, 1, 0, I18n.text("Collapse"), I18n.text("""
                <html><body>
                <b>Roll vs. HT</b> every second to avoid falling unconscious<br>
                Move and Dodge are halved (B419)
                </body></html>"""), ops));
        thresholds.add(new PoolThreshold(1, 3, 0, I18n.text("Reeling"), I18n.text("Move and Dodge are halved (B419)"), ops));
        thresholds.add(new PoolThreshold(1, 1, -1, I18n.text("Wounded"), "", null));
        thresholds.add(new PoolThreshold(1, 1, 0, I18n.text("Healthy"), "", null));
        m.put("hp", new AttributeDef("hp", I18n.text("HP"), I18n.text("Hit Points"), "$st", ++i, 2, 10, thresholds));

        return m;
    }

    public AttributeDef(String id, AttributeType type, String name, String fullName, String attributeBase, int order, int costPerPoint, int costAdjPercentPerSM) {
        setID(id);
        mType = type;
        mName = name;
        mFullName = fullName;
        mAttributeBase = attributeBase;
        mOrder = order;
        mCostPerPoint = costPerPoint;
        mCostAdjPercentPerSM = costAdjPercentPerSM;
    }

    public AttributeDef(String id, String name, String fullName, String attributeBase, int order, int costPerPoint, int costAdjPercentPerSM, List<PoolThreshold> thresholds) {
        setID(id);
        mType = AttributeType.POOL;
        mName = name;
        mFullName = fullName;
        mAttributeBase = attributeBase;
        mOrder = order;
        mCostPerPoint = costPerPoint;
        mCostAdjPercentPerSM = costAdjPercentPerSM;
        mThresholds = thresholds;
    }

    public AttributeDef(JsonMap m, int order) {
        setID(m.getString(KEY_ID));
        mType = Enums.extract(m.getString(KEY_TYPE), AttributeType.values(), AttributeType.INTEGER);
        mName = m.getString(KEY_SHORT_NAME);
        mFullName = m.getString(KEY_FULL_NAME);
        mAttributeBase = m.getString(KEY_ATTRIBUTE_BASE);
        mOrder = order;
        mCostPerPoint = m.getInt(KEY_COST_PER_POINT);
        mCostAdjPercentPerSM = m.getInt(KEY_COST_ADJ_PERCENT_PER_SM);
        if (mType == AttributeType.POOL) {
            JsonArray a      = m.getArray(KEY_THRESHOLDS);
            int       length = a.size();
            mThresholds = new ArrayList<>();
            for (int i = 0; i < length; i++) {
                mThresholds.add(new PoolThreshold(a.getMap(i)));
            }
        }
    }

    public String getID() {
        return mID;
    }

    public void setID(String id) {
        mID = ID.sanitize(id, RESERVED, false);
    }

    public AttributeType getType() {
        return mType;
    }

    public void setType(AttributeType type) {
        mType = type;
        if (mType == AttributeType.POOL) {
            if (mThresholds == null) {
                mThresholds = new ArrayList<>();
            }
        } else {
            mThresholds = null;
        }
    }

    public String getCombinedName() {
        String combinedName = mFullName;
        if (combinedName.isBlank()) {
            combinedName = mName;
        } else if (!mName.isBlank() && !combinedName.equals(mName)) {
            combinedName = String.format("%s (%s)", combinedName, mName);
        }
        return combinedName;
    }

    public String getName() {
        return mName;
    }

    public void setName(String name) {
        mName = name;
    }

    public String getFullName() {
        return mFullName;
    }

    public void setFullName(String fullName) {
        mFullName = fullName;
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

    public List<PoolThreshold> getThresholds() {
        return mThresholds;
    }

    public void setThresholds(List<PoolThreshold> thresholds) {
        mThresholds = thresholds;
    }

    public boolean isPrimary() {
        try {
            Integer.parseInt(mAttributeBase);
            return true;
        } catch (NumberFormatException ex) {
            return false;
        }
    }

    public double getBaseValue(VariableResolver resolver) {
        try {
            return new Evaluator(resolver).evaluateToNumber(mAttributeBase);
        } catch (EvaluationException ex) {
            Log.error(ex);
            return 0;
        }
    }

    public int computeCost(GURPSCharacter character, double value, int sm, int costReduction) {
        if (sm > 0 && mCostAdjPercentPerSM > 0 && !("hp".equals(mID) && character.getSheetSettings().useKnowYourOwnStrength())) {
            costReduction += sm * mCostAdjPercentPerSM;
            if (costReduction < 0) {
                costReduction = 0;
            } else if (costReduction > 80) {
                costReduction = 80;
            }
        }
        int cost = (int) (mCostPerPoint * value);
        if (costReduction != 0) {
            cost *= 100 - costReduction;
            int rem = cost % 100;
            cost /= 100;
            if (rem > 49) {
                cost++;
            } else if (rem < -50) {
                cost--;
            }
        }
        return cost;
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_ID, mID);
        w.keyValue(KEY_TYPE, Enums.toId(mType));
        w.keyValue(KEY_SHORT_NAME, mName);
        w.keyValue(KEY_FULL_NAME, mFullName);
        w.keyValue(KEY_ATTRIBUTE_BASE, mAttributeBase);
        w.keyValue(KEY_COST_PER_POINT, mCostPerPoint);
        w.keyValue(KEY_COST_ADJ_PERCENT_PER_SM, mCostAdjPercentPerSM);
        if (mType == AttributeType.POOL) {
            w.key(KEY_THRESHOLDS);
            w.startArray();
            if (mThresholds != null) {
                for (PoolThreshold threshold : mThresholds) {
                    threshold.toJSON(w);
                }
            }
            w.endArray();
        }
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
        if (mType != that.mType) {
            return false;
        }
        if (mCostPerPoint != that.mCostPerPoint) {
            return false;
        }
        if (mCostAdjPercentPerSM != that.mCostAdjPercentPerSM) {
            return false;
        }
        if (mOrder != that.mOrder) {
            return false;
        }
        if (!mID.equals(that.mID)) {
            return false;
        }
        if (!mName.equals(that.mName)) {
            return false;
        }
        if (!mFullName.equals(that.mFullName)) {
            return false;
        }
        if (!mAttributeBase.equals(that.mAttributeBase)) {
            return false;
        }
        return Objects.equals(mThresholds, that.mThresholds);
    }

    @Override
    public int hashCode() {
        int result = mID.hashCode();
        result = 31 * result + mName.hashCode();
        result = 31 * result + mFullName.hashCode();
        result = 31 * result + mAttributeBase.hashCode();
        result = 31 * result + mCostPerPoint;
        result = 31 * result + mCostAdjPercentPerSM;
        result = 31 * result + mOrder;
        result = 31 * result + mType.ordinal();
        result = 31 * result + (mThresholds != null ? mThresholds.hashCode() : 0);
        return result;
    }

    @Override
    protected AttributeDef clone() {
        AttributeDef other = null;
        try {
            other = (AttributeDef) super.clone();
            if (mType == AttributeType.POOL) {
                if (mThresholds != null) {
                    other.mThresholds = PoolThreshold.cloneList(mThresholds);
                } else {
                    other.mThresholds = new ArrayList<>();
                }
            } else {
                mThresholds = null;
            }
        } catch (CloneNotSupportedException e) {
            // This can't happen
            Log.error(e);
            System.exit(1);
        }
        return other;
    }

    @Override
    // Used solely for remembering the desired ordering.
    public int compareTo(AttributeDef other) {
        return Integer.compare(mOrder, other.mOrder);
    }
}
