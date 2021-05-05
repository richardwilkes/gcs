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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Objects;

public class PoolThreshold implements Cloneable {
    private static final String KEY_STATE       = "state";
    private static final String KEY_EXPLANATION = "explanation";
    private static final String KEY_MULTIPLIER  = "multiplier";
    private static final String KEY_DIVISOR     = "divisor";
    private static final String KEY_ADDITION    = "addition";
    private static final String KEY_OPS         = "ops";

    private String             mState;
    private String             mExplanation;
    private int                mMultiplier;
    private int                mDivisor;
    private int                mAddition;
    private List<ThresholdOps> mOps;

    public static final List<PoolThreshold> cloneList(List<PoolThreshold> list) {
        List<PoolThreshold> result = new ArrayList<>();
        for (PoolThreshold threshold : list) {
            result.add(threshold.clone());
        }
        return result;
    }

    public PoolThreshold(int multiplier, int divisor, int addition, String state, String explanation, List<ThresholdOps> ops) {
        mState = state;
        mExplanation = explanation;
        mMultiplier = multiplier;
        mDivisor = divisor;
        mAddition = addition;
        mOps = ops != null ? ops : new ArrayList<>();
    }

    public PoolThreshold(JsonMap m) {
        mState = m.getString(KEY_STATE);
        mExplanation = m.getString(KEY_EXPLANATION);
        mMultiplier = m.getInt(KEY_MULTIPLIER);
        mDivisor = m.getInt(KEY_DIVISOR);
        mAddition = m.getInt(KEY_ADDITION);
        mOps = new ArrayList<>();
        JsonArray a      = m.getArray(KEY_OPS);
        int       length = a.size();
        for (int i = 0; i < length; i++) {
            mOps.add(Enums.extract(a.getString(i), ThresholdOps.values(), ThresholdOps.UNKNOWN));
        }
    }

    public int threshold(int max) {
        int threshold = max * mMultiplier;
        if (mDivisor > 1) {
            threshold /= mDivisor;
            if (max % mDivisor != 0) {
                threshold++;
            }
            if (--threshold < 0) {
                threshold = 0;
            }
        }
        return threshold + mAddition;
    }

    public int getMultiplier() {
        return mMultiplier;
    }

    public void setMultiplier(int multiplier) {
        mMultiplier = multiplier;
    }

    public int getDivisor() {
        return mDivisor;
    }

    public void setDivisor(int divisor) {
        mDivisor = divisor;
    }

    public int getAddition() {
        return mAddition;
    }

    public void setAddition(int addition) {
        mAddition = addition;
    }

    public String getState() {
        return mState;
    }

    public void setState(String state) {
        mState = state;
    }

    public String getExplanation() {
        return mExplanation;
    }

    public void setExplanation(String explanation) {
        mExplanation = explanation;
    }

    public List<ThresholdOps> getOps() {
        return mOps;
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_STATE, mState);
        w.keyValue(KEY_EXPLANATION, mExplanation);
        w.keyValue(KEY_MULTIPLIER, mMultiplier);
        w.keyValue(KEY_DIVISOR, mDivisor);
        w.keyValue(KEY_ADDITION, mAddition);
        w.key(KEY_OPS);
        w.startArray();
        Collections.sort(mOps);
        for (ThresholdOps op : mOps) {
            w.value(Enums.toId(op));
        }
        w.endArray();
        w.endMap();
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (!(o instanceof PoolThreshold)) {
            return false;
        }
        PoolThreshold that = (PoolThreshold) o;
        if (mMultiplier != that.mMultiplier) {
            return false;
        }
        if (mDivisor != that.mDivisor) {
            return false;
        }
        if (mAddition != that.mAddition) {
            return false;
        }
        if (!mState.equals(that.mState)) {
            return false;
        }
        if (!mExplanation.equals(that.mExplanation)) {
            return false;
        }
        return Objects.equals(mOps, that.mOps);
    }

    @Override
    public int hashCode() {
        int result = mState.hashCode();
        result = 31 * result + mExplanation.hashCode();
        result = 31 * result + mMultiplier;
        result = 31 * result + mDivisor;
        result = 31 * result + mAddition;
        result = 31 * result + mOps.hashCode();
        return result;
    }

    @Override
    protected PoolThreshold clone() {
        PoolThreshold other = null;
        try {
            other = (PoolThreshold) super.clone();
            other.mOps = new ArrayList<>(mOps);
        } catch (CloneNotSupportedException e) {
            // This can't happen
            Log.error(e);
            System.exit(1);
        }
        return other;
    }
}
