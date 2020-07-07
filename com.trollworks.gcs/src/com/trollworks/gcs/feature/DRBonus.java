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

import com.trollworks.gcs.character.Armor;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.io.IOException;

/** A DR bonus. */
public class DRBonus extends Bonus {
    /** The XML tag. */
    public static final  String      TAG_ROOT     = "dr_bonus";
    private static final String      TAG_LOCATION = "location";
    private              HitLocation mLocation;

    /** Creates a new DR bonus. */
    public DRBonus() {
        super(1);
        mLocation = HitLocation.TORSO;
    }

    public DRBonus(JsonMap m) throws IOException {
        this();
        loadSelf(m);
    }

    /**
     * Loads a {@link DRBonus}.
     *
     * @param reader The XML reader to use.
     */
    public DRBonus(XMLReader reader) throws IOException {
        this();
        load(reader);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public DRBonus(DRBonus other) {
        super(other);
        mLocation = other.mLocation;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof DRBonus && super.equals(obj)) {
            return mLocation == ((DRBonus) obj).mLocation;
        }
        return false;
    }

    @Override
    public String getJSONTypeName() {
        return TAG_ROOT;
    }

    @Override
    public String getXMLTag() {
        return TAG_ROOT;
    }

    @Override
    public String getKey() {
        return Armor.DR_PREFIX + mLocation.name();
    }

    @Override
    public Feature cloneFeature() {
        return new DRBonus(this);
    }

    @Override
    protected void loadSelf(XMLReader reader) throws IOException {
        if (TAG_LOCATION.equals(reader.getName())) {
            setLocation(Enums.extract(reader.readText(), HitLocation.values(), HitLocation.TORSO));
        } else {
            super.loadSelf(reader);
        }
    }

    @Override
    protected void loadSelf(JsonMap m) throws IOException {
        super.loadSelf(m);
        setLocation(Enums.extract(m.getString(TAG_LOCATION), HitLocation.values(), HitLocation.TORSO));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(TAG_LOCATION, Enums.toId(mLocation));
    }

    /** @return The location protected by the DR. */
    public HitLocation getLocation() {
        return mLocation;
    }

    /** @param location The location. */
    public void setLocation(HitLocation location) {
        mLocation = location;
    }
}
