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

package com.trollworks.gcs.criteria;

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.units.WeightValue;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** Manages weight comparison criteria. */
public class WeightCriteria extends NumericCriteria {
    private WeightValue mQualifier;

    /**
     * Creates a new WeightCriteria.
     *
     * @param type      The {@link NumericCompareType} to use.
     * @param qualifier The qualifier to match against.
     */
    public WeightCriteria(NumericCompareType type, WeightValue qualifier) {
        super(type);
        setQualifier(qualifier);
    }

    /**
     * Creates a new WeightCriteria.
     *
     * @param other A {@link WeightCriteria} to clone.
     */
    public WeightCriteria(WeightCriteria other) {
        super(other.getType());
        mQualifier = new WeightValue(other.mQualifier);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof WeightCriteria && super.equals(obj)) {
            return mQualifier.equals(((WeightCriteria) obj).mQualifier);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public void load(XMLReader reader) throws IOException {
        super.load(reader);
        setQualifier(WeightValue.extract(reader.readText(), false));
    }

    @Override
    public void load(JsonMap m) throws IOException {
        super.load(m);
        setQualifier(WeightValue.extract(m.getString(KEY_QUALIFIER), false));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_QUALIFIER, mQualifier.toString(false));
    }

    /** @return The qualifier to match against. */
    public WeightValue getQualifier() {
        return mQualifier;
    }

    @Override
    public String getQualifierAsString(boolean allowAdornments) {
        return mQualifier.toString(allowAdornments);
    }

    /** @param qualifier The qualifier to match against. */
    public void setQualifier(WeightValue qualifier) {
        mQualifier = new WeightValue(qualifier);
    }

    /**
     * @param data The data to match against.
     * @return Whether the data matches this criteria.
     */
    public boolean matches(WeightValue data) {
        switch (getType()) {
        case IS:
            return mQualifier.equals(data);
        case AT_LEAST:
        default:
            return data.getNormalizedValue().greaterThanOrEqual(mQualifier.getNormalizedValue());
        case AT_MOST:
            return data.getNormalizedValue().lessThanOrEqual(mQualifier.getNormalizedValue());
        }
    }
}
