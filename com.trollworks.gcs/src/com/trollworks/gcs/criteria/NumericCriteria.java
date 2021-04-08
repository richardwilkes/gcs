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

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.IOException;

/** Manages numeric comparison criteria. */
public abstract class NumericCriteria extends Criteria {
    private NumericCompareType mType;

    /**
     * Creates a new numeric comparison.
     *
     * @param type The {@link NumericCompareType} to use.
     */
    public NumericCriteria(NumericCompareType type) {
        setType(type);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof NumericCriteria) {
            return mType == ((NumericCriteria) obj).mType;
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public void load(JsonMap m) throws IOException {
        setType(Enums.extract(m.getString(ATTRIBUTE_COMPARE), NumericCompareType.values(), NumericCompareType.AT_LEAST));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.keyValue(ATTRIBUTE_COMPARE, Enums.toId(mType));
    }

    /**
     * @param allowAdornments Whether extras, such as "lbs." can be appended to the text.
     * @return The numeric qualifier, as a {@link String}.
     */
    public abstract String getQualifierAsString(boolean allowAdornments);

    /** @return The type of comparison to make. */
    public NumericCompareType getType() {
        return mType;
    }

    /** @param type The type of comparison to make. */
    public void setType(NumericCompareType type) {
        mType = type;
    }

    @Override
    public String toString() {
        return toString(I18n.Text("is "));
    }

    /**
     * @param prefix A prefix to place before the description.
     * @return A formatted description of this object.
     */
    public String toString(String prefix) {
        return mType.format(prefix, getQualifierAsString(true));
    }
}
