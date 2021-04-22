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

package com.trollworks.gcs.pointpool;

import com.trollworks.gcs.attribute.Attribute;
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
import java.util.Objects;

public class PointPoolDef implements Cloneable, Comparable<PointPoolDef> {
    private static final String KEY_ID                      = "id";
    private static final String KEY_NAME                    = "name";
    private static final String KEY_DESCRIPTION             = "description";
    private static final String KEY_ATTRIBUTE_BASE          = "attribute_base";
    private static final String KEY_COST_PER_POINT          = "cost_per_point";
    private static final String KEY_COST_ADJ_PERCENT_PER_SM = "cost_adj_percent_per_sm";
    private static final String KEY_THRESHOLDS              = "thresholds";

    private String              mID;
    private String              mName;
    private String              mDescription;
    private String              mAttributeBase;
    private int                 mOrder;
    private int                 mCostPerPoint;
    private int                 mCostAdjPercentPerSM;
    private List<PoolThreshold> mThresholds;

    public static final Map<String, PointPoolDef> cloneMap(Map<String, PointPoolDef> m) {
        Map<String, PointPoolDef> result = new HashMap<>();
        for (PointPoolDef pool : m.values()) {
            result.put(pool.getID(), pool.clone());
        }
        return result;
    }

    public static final Map<String, PointPoolDef> loadPools(JsonArray a) {
        Map<String, PointPoolDef> p      = new HashMap<>();
        int                       length = a.size();
        for (int i = 0; i < length; i++) {
            PointPoolDef pool = new PointPoolDef(a.getMap(i), i);
            p.put(pool.getID(), pool);
        }
        return p;
    }

    public static final List<PointPoolDef> getOrderedPools(Map<String, PointPoolDef> m) {
        List<PointPoolDef> ordered = new ArrayList<>(m.values());
        Collections.sort(ordered);
        return ordered;
    }

    public static final void writeOrderedPools(JsonWriter w, Map<String, PointPoolDef> m) throws IOException {
        w.startArray();
        for (PointPoolDef p : getOrderedPools(m)) {
            p.toJSON(w);
        }
        w.endArray();

    }

    public static final Map<String, PointPoolDef> createStandardPools() {
        List<ThresholdOps> ops = new ArrayList<>();
        ops.add(ThresholdOps.HALVE_MOVE);
        ops.add(ThresholdOps.HALVE_DODGE);
        ops.add(ThresholdOps.HALVE_ST);

        List<PoolThreshold> thresholds = new ArrayList<>();
        thresholds.add(new PoolThreshold(-1, 1, 0, I18n.Text("Unconscious"), "", ops));
        thresholds.add(new PoolThreshold(0, 1, 0, I18n.Text("Collapse"), I18n.Text("""
                <html><body>
                <b>Roll vs. Will</b> to do anything besides talk or rest; failure causes unconsciousness<br>
                Each FP you lose below 0 also causes 1 HP of injury<br>
                Move, Dodge and ST are halved (B426)
                </body></html>
                """), ops));
        thresholds.add(new PoolThreshold(1, 3, 0, I18n.Text("Tired"), I18n.Text("Move, Dodge and ST are halved (B426)"), ops));
        thresholds.add(new PoolThreshold(1, 1, -1, I18n.Text("Tiring"), "", null));
        thresholds.add(new PoolThreshold(1, 1, 0, I18n.Text("Rested"), "", null));

        Map<String, PointPoolDef> m = new HashMap<>();
        m.put("fp", new PointPoolDef("fp", I18n.Text("FP"), I18n.Text("Fatigue Points"), "$" + Attribute.ID_ATTR_PREFIX + "ht", 0, 3, 0, thresholds));

        ops = new ArrayList<>();
        ops.add(ThresholdOps.HALVE_MOVE);
        ops.add(ThresholdOps.HALVE_DODGE);

        thresholds = new ArrayList<>();
        thresholds.add(new PoolThreshold(-5, 1, 0, I18n.Text("Dead"), "", ops));
        for (int i = -4; i < 0; i++) {
            thresholds.add(new PoolThreshold(i, 1, 0, String.format(I18n.Text("Dying #%d"), Integer.valueOf(-i)), String.format(I18n.Text("""
                    <html><body>
                    <b>Roll vs. HT</b> to avoid death<br>
                    <b>Roll vs. HT%d</b> every second to avoid falling unconscious<br>
                    Move and Dodge are halved (B419)
                    </body></html>
                    """), Integer.valueOf(i)), ops));
        }
        thresholds.add(new PoolThreshold(0, 1, 0, I18n.Text("Collapse"), I18n.Text("""
                <html><body>
                <b>Roll vs. HT</b> every second to avoid falling unconscious<br>
                Move and Dodge are halved (B419)
                </body></html>
                """), ops));
        thresholds.add(new PoolThreshold(1, 3, 0, I18n.Text("Reeling"), I18n.Text("Move and Dodge are halved (B419)"), ops));
        thresholds.add(new PoolThreshold(1, 1, -1, I18n.Text("Wounded"), "", null));
        thresholds.add(new PoolThreshold(1, 1, 0, I18n.Text("Healthy"), "", null));
        m.put("hp", new PointPoolDef("hp", I18n.Text("HP"), I18n.Text("Hit Points"), "$" + Attribute.ID_ATTR_PREFIX + "st", 1, 2, 10, thresholds));
        return m;
    }

    public PointPoolDef(String id, String name, String desc, String attributeBase, int order, int costPerPoint, int costAdjPercentPerSM, List<PoolThreshold> thresholds) {
        mID = id;
        mName = name;
        mDescription = desc;
        mAttributeBase = attributeBase;
        mOrder = order;
        mCostPerPoint = costPerPoint;
        mCostAdjPercentPerSM = costAdjPercentPerSM;
        mThresholds = thresholds;
    }

    public PointPoolDef(JsonMap m, int order) {
        mID = m.getString(KEY_ID);
        mName = m.getString(KEY_NAME);
        mDescription = m.getString(KEY_DESCRIPTION);
        mAttributeBase = m.getString(KEY_ATTRIBUTE_BASE);
        mOrder = order;
        mCostPerPoint = m.getInt(KEY_COST_PER_POINT);
        mCostAdjPercentPerSM = m.getInt(KEY_COST_ADJ_PERCENT_PER_SM);
        mThresholds = new ArrayList<>();
        if (m.has(KEY_THRESHOLDS)) {
            JsonArray a      = m.getArray(KEY_THRESHOLDS);
            int       length = a.size();
            for (int i = 0; i < length; i++) {
                mThresholds.add(new PoolThreshold(a.getMap(i)));
            }
        }
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

    public List<PoolThreshold> getThresholds() {
        return mThresholds;
    }

    public void setThresholds(List<PoolThreshold> thresholds) {
        mThresholds = thresholds;
    }

    public int getBaseValue(CharacterVariableResolver resolver) {
        Evaluator evaluator = new Evaluator(resolver);
        String    exclude   = PointPool.ID_POOL_PREFIX + mID;
        resolver.addExclusion(exclude);
        int value;
        try {
            value = evaluator.evaluateToInteger(mAttributeBase);
        } catch (EvaluationException ex) {
            Log.error(ex);
            value = 0;
        }
        resolver.removeExclusion(exclude);
        return value;
    }

    public int computeCost(GURPSCharacter character, int value, int sm) {
        int cost = mCostPerPoint * value;
        if (sm > 0 && mCostAdjPercentPerSM > 0 && !("hp".equals(mID) && character.getSettings().useKnowYourOwnStrength())) {
            int factor = sm * mCostAdjPercentPerSM;
            if (factor > 80) {
                factor = 80;
            }
            cost *= 100 - factor;
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
        w.keyValue(KEY_NAME, mName);
        w.keyValue(KEY_DESCRIPTION, mDescription);
        w.keyValue(KEY_ATTRIBUTE_BASE, mAttributeBase);
        w.keyValue(KEY_COST_PER_POINT, mCostPerPoint);
        w.keyValue(KEY_COST_ADJ_PERCENT_PER_SM, mCostAdjPercentPerSM);
        w.key(KEY_THRESHOLDS);
        w.startArray();
        for (PoolThreshold threshold : mThresholds) {
            threshold.toJSON(w);
        }
        w.endArray();
        w.endMap();
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (!(o instanceof PointPoolDef)) {
            return false;
        }
        PointPoolDef that = (PointPoolDef) o;
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
        if (!mAttributeBase.equals(that.mAttributeBase)) {
            return false;
        }
        return Objects.equals(mThresholds, that.mThresholds);
    }

    @Override
    public int hashCode() {
        int result = mID.hashCode();
        result = 31 * result + mName.hashCode();
        result = 31 * result + mDescription.hashCode();
        result = 31 * result + mAttributeBase.hashCode();
        result = 31 * result + mCostPerPoint;
        result = 31 * result + mCostAdjPercentPerSM;
        result = 31 * result + (mThresholds != null ? mThresholds.hashCode() : 0);
        return result;
    }

    @Override
    protected PointPoolDef clone() {
        PointPoolDef other = null;
        try {
            other = (PointPoolDef) super.clone();
            if (mThresholds != null) {
                other.mThresholds = PoolThreshold.cloneList(mThresholds);
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
    public int compareTo(PointPoolDef other) {
        return Integer.compare(mOrder, other.mOrder);
    }
}
