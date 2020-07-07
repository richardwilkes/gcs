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
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** Manages string comparison criteria. */
public class StringCriteria extends Criteria {
    private StringCompareType mType;
    private String            mQualifier;

    /**
     * Creates a new string comparison.
     *
     * @param type      The type of comparison.
     * @param qualifier The qualifier to match against.
     */
    public StringCriteria(StringCompareType type, String qualifier) {
        setType(type);
        setQualifier(qualifier);
    }

    /**
     * Creates a new string comparison.
     *
     * @param other A {@link StringCriteria} to clone.
     */
    public StringCriteria(StringCriteria other) {
        mType = other.mType;
        mQualifier = other.mQualifier;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof StringCriteria) {
            StringCriteria sc = (StringCriteria) obj;
            return mType == sc.mType && mQualifier.equalsIgnoreCase(sc.mQualifier);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public void load(JsonMap m) throws IOException {
        setQualifier(m.getString(KEY_QUALIFIER));
        setType(Enums.extract(m.getString(ATTRIBUTE_COMPARE), StringCompareType.values(), StringCompareType.ANY));
    }

    @Override
    public void save(JsonWriter w, String key) throws IOException {
        if (!isTypeAnything()) {
            super.save(w, key);
        }
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        w.keyValue(ATTRIBUTE_COMPARE, Enums.toId(mType));
        w.keyValueNot(KEY_QUALIFIER, mQualifier, "");
    }

    /**
     * @param reader The reader to load data from.
     */
    public void load(XMLReader reader) throws IOException {
        setType(Enums.extract(reader.getAttribute(ATTRIBUTE_COMPARE), StringCompareType.values(), StringCompareType.ANY));
        setQualifier(reader.readText());
    }

    /** @return The type of comparison to make. */
    public StringCompareType getType() {
        return mType;
    }

    /** @param type The type of comparison to make. */
    public void setType(StringCompareType type) {
        mType = type;
    }

    /** @return The qualifier to match against. */
    public String getQualifier() {
        return mQualifier;
    }

    /** @param qualifier The qualifier to match against. */
    public void setQualifier(String qualifier) {
        mQualifier = qualifier != null ? qualifier : "";
    }

    /**
     * @param data The data to match against.
     * @return Whether the data matches this criteria.
     */
    public boolean matches(String data) {
        return mType.matches(mQualifier, data);
    }

    @Override
    public String toString() {
        return mType.describe(mQualifier);
    }

    /** @return Is this criteria for an exact match? */
    public boolean isTypeIs() {
        return mType == StringCompareType.IS;
    }

    /** @return Is this criteria for any match? */
    public boolean isTypeAnything() {
        return mType == StringCompareType.ANY;
    }
}
