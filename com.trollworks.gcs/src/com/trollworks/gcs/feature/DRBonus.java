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

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

/** A DR bonus. */
public class DRBonus extends Bonus {
    public static final  String ALL_SPECIALIZATION = "all";
    public static final  String KEY_ROOT           = "dr_bonus";
    private static final String KEY_LOCATION       = "location";
    private static final String KEY_SPECIALIZATION = "specialization";

    private String mLocation;
    private String mSpecialization;

    /** Creates a new DR bonus. */
    public DRBonus() {
        super(1);
        mLocation = "torso";
        mSpecialization = ALL_SPECIALIZATION;
    }

    public DRBonus(DataFile dataFile, JsonMap m) throws IOException {
        this();
        loadSelf(dataFile, m);
    }

    /**
     * Creates a clone of the specified bonus.
     *
     * @param other The bonus to clone.
     */
    public DRBonus(DRBonus other) {
        super(other);
        mLocation = other.mLocation;
        mSpecialization = other.mSpecialization;
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof DRBonus other && super.equals(obj)) {
            return mLocation.equals(other.mLocation) && mSpecialization.equals(other.mSpecialization);
        }
        return false;
    }

    @Override
    public String getJSONTypeName() {
        return KEY_ROOT;
    }

    @Override
    public String getKey() {
        return com.trollworks.gcs.body.HitLocation.KEY_PREFIX + mLocation;
    }

    @Override
    public Feature cloneFeature() {
        return new DRBonus(this);
    }

    @Override
    protected void loadSelf(DataFile dataFile, JsonMap m) throws IOException {
        super.loadSelf(dataFile, m);
        mLocation = m.getString(KEY_LOCATION);
        setSpecialization(m.getString(KEY_SPECIALIZATION));
    }

    @Override
    protected void saveSelf(JsonWriter w) throws IOException {
        super.saveSelf(w);
        w.keyValue(KEY_LOCATION, mLocation);
        w.keyValueNot(KEY_SPECIALIZATION, mSpecialization, ALL_SPECIALIZATION);
    }

    /** @return The location protected by the DR. */
    public String getLocation() {
        return mLocation;
    }

    /** @param location The location. */
    public void setLocation(String location) {
        mLocation = location;
    }

    /** @return The specialization. */
    public String getSpecialization() {
        return mSpecialization;
    }

    /** @param specialization The specialization. */
    public void setSpecialization(String specialization) {
        specialization = specialization.trim();
        if (specialization.isBlank() || ALL_SPECIALIZATION.equalsIgnoreCase(specialization)) {
            mSpecialization = ALL_SPECIALIZATION;
        } else {
            mSpecialization = specialization;
        }
    }

    @Override
    public void addToToolTip(StringBuilder toolTip) {
        if (toolTip != null) {
            toolTip.append("\n").append(getParentName()).append(" [").append(getToolTipAmount()).append(" against ").append(mSpecialization).append(" attacks]");
        }
    }
}
