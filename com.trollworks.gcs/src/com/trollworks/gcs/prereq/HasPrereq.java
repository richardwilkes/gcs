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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** An abstract prerequisite class for whether or not the specific item is present. */
public abstract class HasPrereq extends Prereq {
    /** The "has" attribute name. */
    protected static final String  ATTRIBUTE_HAS = "has";
    private                boolean mHas;

    /**
     * Creates a new prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     */
    public HasPrereq(PrereqList parent) {
        super(parent);
        mHas = true;
    }

    /**
     * Creates a copy of the specified prerequisite.
     *
     * @param parent The owning prerequisite list, if any.
     * @param prereq The prerequisite to clone.
     */
    protected HasPrereq(PrereqList parent, HasPrereq prereq) {
        super(parent);
        mHas = prereq.mHas;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof HasPrereq) {
            return mHas == ((HasPrereq) obj).mHas;
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }

    @Override
    public void loadSelf(JsonMap m, LoadState state) throws IOException {
        mHas = m.getBoolean(ATTRIBUTE_HAS);
    }

    @Override
    public void saveSelf(JsonWriter w) throws IOException {
        w.keyValue(ATTRIBUTE_HAS, mHas);
    }

    /**
     * Loads the "has" attribute.
     *
     * @param reader The XML reader to load from.
     */
    protected void loadHasAttribute(XMLReader reader) {
        mHas = reader.isAttributeSet(ATTRIBUTE_HAS);
    }

    /**
     * @return {@code true} if the specified criteria should exist, {@code false} if it should not.
     */
    public boolean has() {
        return mHas;
    }

    /**
     * @param has {@code true} if the specified criteria should exist, {@code false} if it should
     *            not.
     */
    public void has(boolean has) {
        mHas = has;
    }

    /** @return The text associated with the current has() state. */
    public String hasText() {
        return mHas ? I18n.Text("Has") : I18n.Text("Does not have");
    }
}
