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

package com.trollworks.gcs.character;

/** An entry in a hit location table. */
public class HitLocationTableEntry {
    private HitLocation mLocation;
    private String      mTitleOverride;
    private int         mHitPenaltyModifier;
    private int         mLowRoll;
    private int         mHighRoll;

    public HitLocationTableEntry(HitLocation location) {
        this(location, null, 0, 0, 0);
    }

    public HitLocationTableEntry(HitLocation location, int hitPenaltyModifier) {
        this(location, null, hitPenaltyModifier, 0, 0);
    }

    public HitLocationTableEntry(HitLocation location, int lowRoll, int highRoll) {
        this(location, null, 0, lowRoll, highRoll);
    }

    public HitLocationTableEntry(HitLocation location, String titleOverride, int lowRoll, int highRoll) {
        this(location, titleOverride, 0, lowRoll, highRoll);
    }

    public HitLocationTableEntry(HitLocation location, String titleOverride, int hitPenaltyModifier, int lowRoll, int highRoll) {
        mLocation = location;
        mTitleOverride = titleOverride;
        mHitPenaltyModifier = hitPenaltyModifier;
        mLowRoll = lowRoll;
        mHighRoll = highRoll;
    }

    public String getKey() {
        return mLocation.getKey();
    }

    public HitLocation getLocation() {
        return mLocation;
    }

    public String getName() {
        return mTitleOverride != null ? mTitleOverride : mLocation.getTitle();
    }

    public int getHitPenalty() {
        return mLocation.getHitPenalty() + mHitPenaltyModifier;
    }

    public boolean hasRoll() {
        return mLowRoll != 0;
    }

    public String getRoll() {
        if (hasRoll()) {
            if (mLowRoll == mHighRoll) {
                return Integer.toString(mLowRoll);
            }
            return mLowRoll + "-" + mHighRoll;
        }
        return "-";
    }
}
