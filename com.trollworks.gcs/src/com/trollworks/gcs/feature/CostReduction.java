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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.attribute.Attribute;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.Set;

/** Describes a cost reduction. */
public class CostReduction extends Feature {
    public static final  String KEY_ROOT       = "cost_reduction";
    private static final String KEY_ATTRIBUTE  = "attribute";
    private static final String KEY_PERCENTAGE = "percentage";

    private String mAttribute;
    private int    mPercentage;

    /** Creates a new cost reduction. */
    public CostReduction() {
        List<AttributeDef> list = AttributeDef.getOrdered(Preferences.getInstance().getSheetSettings().getAttributes());
        mAttribute = list.isEmpty() ? "st" : list.get(0).getID();
        mPercentage = 40;
    }

    /**
     * Creates a clone of the specified cost reduction.
     *
     * @param other The bonus to clone.
     */
    public CostReduction(CostReduction other) {
        mAttribute = other.mAttribute;
        mPercentage = other.mPercentage;
    }

    public CostReduction(JsonMap m) {
        this();
        load(m);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof CostReduction) {
            CostReduction cr = (CostReduction) obj;
            return mPercentage == cr.mPercentage && mAttribute.equals(cr.mAttribute);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    /** @return The percentage to use. */
    public int getPercentage() {
        return mPercentage;
    }

    /** @param percentage The percentage to use. */
    public void setPercentage(int percentage) {
        mPercentage = percentage;
    }

    /** @return The attribute this cost reduction applies to. */
    public String getAttribute() {
        return mAttribute;
    }

    /** @param attribute The attribute. */
    public void setAttribute(String attribute) {
        mAttribute = attribute;
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public String getKey() {
        return Attribute.ID_ATTR_PREFIX + mAttribute;
    }

    @Override
    public Feature cloneFeature() {
        return new CostReduction(this);
    }

    protected void load(JsonMap m) {
        mAttribute = m.getString(KEY_ATTRIBUTE);
        mPercentage = m.getInt(KEY_PERCENTAGE);
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.keyValue(KEY_ATTRIBUTE, mAttribute);
        w.keyValue(KEY_PERCENTAGE, mPercentage);
    }

    @Override
    public void fillWithNameableKeys(Set<String> set) {
        // Nothing to do.
    }

    @Override
    public void applyNameableKeys(Map<String, String> map) {
        // Nothing to do.
    }
}
