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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.IOException;

/** An attribute bonus. */
public class AttributeBonus extends Bonus {
    /** The XML tag. */
    public static final  String                   TAG_ROOT             = "attribute_bonus";
    private static final String                   TAG_ATTRIBUTE        = "attribute";
    private static final String                   ATTRIBUTE_LIMITATION = "limitation";
    private              BonusAttributeType       mAttribute;
    private              AttributeBonusLimitation mLimitation;

    /** Creates a new attribute bonus. */
    public AttributeBonus() {
        super(1);
        mAttribute = BonusAttributeType.ST;
        mLimitation = AttributeBonusLimitation.NONE;
    }

    public AttributeBonus(JsonMap m) throws IOException {
        this();
        loadSelf(m);
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
            return mAttribute == ab.mAttribute && mLimitation == ab.mLimitation;
        }
        return false;
    }

    @Override
    public Feature cloneFeature() {
        return new AttributeBonus(this);
    }

    @Override
    public String getJSONTypeName() {
        return TAG_ROOT;
    }

    @Override
    public String getKey() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(GURPSCharacter.ATTRIBUTES_PREFIX);
        buffer.append(mAttribute.name());
        if (mLimitation != AttributeBonusLimitation.NONE) {
            buffer.append(mLimitation.name());
        }
        return buffer.toString();
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        setAttribute(Enums.extract(m.getString(TAG_ATTRIBUTE), BonusAttributeType.values(), BonusAttributeType.ST));
        setLimitation(Enums.extract(m.getString(ATTRIBUTE_LIMITATION), AttributeBonusLimitation.values(), AttributeBonusLimitation.NONE));
        super.loadSelf(m);
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(TAG_ATTRIBUTE, Enums.toId(mAttribute));
        if (mLimitation != AttributeBonusLimitation.NONE) {
            w.keyValue(ATTRIBUTE_LIMITATION, Enums.toId(mLimitation));
        }
    }

    /** @return The attribute this bonus applies to. */
    public BonusAttributeType getAttribute() {
        return mAttribute;
    }

    /** @param attribute The attribute. */
    public void setAttribute(BonusAttributeType attribute) {
        mAttribute = attribute;
        getAmount().setIntegerOnly(mAttribute.isIntegerOnly());
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
