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
import com.trollworks.gcs.attribute.AttributeType;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.IOException;
import java.util.List;

/** An attribute bonus. */
public class AttributeBonus extends Bonus {
    public static final  String KEY_ROOT       = "attribute_bonus";
    private static final String KEY_ATTRIBUTE  = "attribute";
    private static final String KEY_LIMITATION = "limitation";

    private String                   mAttribute;
    private AttributeBonusLimitation mLimitation;

    /** Creates a new attribute bonus. */
    public AttributeBonus() {
        super(1);
        List<AttributeDef> list = AttributeDef.getOrdered(Preferences.getInstance().getAttributes());
        mAttribute = list.isEmpty() ? "st" : list.get(0).getID();
        mLimitation = AttributeBonusLimitation.NONE;
    }

    public AttributeBonus(DataFile dataFile, JsonMap m) throws IOException {
        this();
        loadSelf(dataFile, m);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public AttributeBonus(AttributeBonus other) {
        super(other);
        mAttribute = other.mAttribute;
        mLimitation = other.mLimitation;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof AttributeBonus && super.equals(obj)) {
            AttributeBonus ab = (AttributeBonus) obj;
            return mAttribute.equals(ab.mAttribute) && mLimitation == ab.mLimitation;
        }
        return false;
    }

    @Override
    public Feature cloneFeature() {
        return new AttributeBonus(this);
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public String getKey() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(Attribute.ID_ATTR_PREFIX);
        buffer.append(mAttribute);
        if (mLimitation != AttributeBonusLimitation.NONE) {
            buffer.append(".");
            buffer.append(mLimitation.name());
        }
        return buffer.toString();
    }

    @Override
    protected void loadSelf(DataFile dataFile, JsonMap m) throws IOException {
        setAttribute(dataFile, m.getString(KEY_ATTRIBUTE));
        mLimitation = Enums.extract(m.getString(KEY_LIMITATION), AttributeBonusLimitation.values(), AttributeBonusLimitation.NONE);
        super.loadSelf(dataFile, m);
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_ATTRIBUTE, mAttribute);
        if (mLimitation != AttributeBonusLimitation.NONE) {
            w.keyValue(KEY_LIMITATION, Enums.toId(mLimitation));
        }
    }

    /** @return The attribute this bonus applies to. */
    public String getAttribute() {
        return mAttribute;
    }

    /** @param attribute The attribute. */
    public void setAttribute(DataFile dataFile, String attribute) {
        mAttribute = attribute;
        AttributeDef def = dataFile.getAttributeDefs().get(attribute);
        getAmount().setIntegerOnly(def == null || def.getType() != AttributeType.DECIMAL);
    }

    /** @return The limitation of this bonus. */
    public AttributeBonusLimitation getLimitation() {
        return mLimitation;
    }

    /** @param limitation The limitation. */
    public void setLimitation(AttributeBonusLimitation limitation) {
        mLimitation = limitation;
    }
}
