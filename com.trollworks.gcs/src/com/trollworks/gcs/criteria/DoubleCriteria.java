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

package com.trollworks.gcs.criteria;

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

/** Manages double comparison criteria. */
public class DoubleCriteria extends NumericCriteria {
    private double mQualifier;

    /**
     * Creates a new double comparison.
     *
     * @param type      The {@link NumericCompareType} to use.
     * @param qualifier The qualifier to match against.
     */
    public DoubleCriteria(NumericCompareType type, double qualifier) {
        super(type);
        setQualifier(qualifier);
    }

    /**
     * Creates a new double comparison.
     *
     * @param other A {@link DoubleCriteria} to clone.
     */
    public DoubleCriteria(DoubleCriteria other) {
        super(other.getType());
        mQualifier = other.mQualifier;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof DoubleCriteria && super.equals(obj)) {
            return mQualifier == ((DoubleCriteria) obj).mQualifier;
        }
        return false;
    }

    @Override
    public void load(JsonMap m) throws IOException {
        super.load(m);
        setQualifier(m.getDouble(KEY_QUALIFIER));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_QUALIFIER, mQualifier);
    }

    /** @return The qualifier to match against. */
    public double getQualifier() {
        return mQualifier;
    }

    @Override
    public String getQualifierAsString(boolean allowAdornments) {
        return Double.toString(mQualifier);
    }

    /** @param qualifier The qualifier to match against. */
    public void setQualifier(double qualifier) {
        mQualifier = qualifier;
    }

    /**
     * @param data The data to match against.
     * @return Whether the data matches this criteria.
     */
    public boolean matches(double data) {
        return switch (getType()) {
            case IS -> data == mQualifier;
            case AT_MOST -> data <= mQualifier;
            default -> data >= mQualifier;
        };
    }
}
