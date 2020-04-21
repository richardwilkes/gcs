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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.utility.text.Numbers;

/** Provides basic timing facilities. */
public final class Timing {
    private long mBase;

    /** Creates a new {@link Timing}. */
    public Timing() {
        reset();
    }

    /** Resets the base time to this instant. */
    public void reset() {
        mBase = System.nanoTime();
    }

    /**
     * @return The number of elapsed nanoseconds since the timing object was created or last reset.
     */
    public long elapsed() {
        return System.nanoTime() - mBase;
    }

    /**
     * @return The number of elapsed nanoseconds since the timing object was created or last reset,
     *         then resets it.
     */
    public long elapsedThenReset() {
        long oldBase = mBase;
        mBase = System.nanoTime();
        return mBase - oldBase;
    }

    /** @return The number of elapsed seconds since the timing object was created or last reset. */
    public double elapsedSeconds() {
        return elapsed() / 1000000000.0;
    }

    /**
     * @return The number of elapsed seconds since the timing object was created or last reset, then
     *         resets it.
     */
    public double elapsedSecondsThenReset() {
        return elapsedThenReset() / 1000000000.0;
    }

    public String toStringWithNanoResolution() {
        return String.format("%,.9fs", Double.valueOf(elapsedSeconds()));
    }

    public String toStringWithNanoResolutionThenReset() {
        return String.format("%,.9fs", Double.valueOf(elapsedSecondsThenReset()));
    }

    public String toStringWithMicroResolution() {
        return String.format("%,.6fs", Double.valueOf(elapsedSeconds()));
    }

    public String toStringWithMicroResolutionThenReset() {
        return String.format("%,.6fs", Double.valueOf(elapsedSecondsThenReset()));
    }

    public String toStringWithMilliResolution() {
        return String.format("%,.3fs", Double.valueOf(elapsedSeconds()));
    }

    public String toStringWithMilliResolutionThenReset() {
        return String.format("%,.3fs", Double.valueOf(elapsedSecondsThenReset()));
    }

    @Override
    public String toString() {
        return Numbers.trimTrailingZeroes(toStringWithMicroResolutionThenReset(), true);
    }
}
