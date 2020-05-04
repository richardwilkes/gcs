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

package com.trollworks.gcs.utility.units;

import com.trollworks.gcs.utility.Fixed6;

/**
 * Holds a value and {@link Units} pair.
 *
 * @param <T> The type of {@link Units} to use.
 */
public class UnitsValue<T extends Units> implements Comparable<UnitsValue<T>> {
    private Fixed6 mValue;
    private T      mUnits;

    /**
     * Creates a new {@link UnitsValue}.
     *
     * @param value The value to use.
     * @param units The {@link Units} to use.
     */
    public UnitsValue(Fixed6 value, T units) {
        mValue = value;
        mUnits = units;
    }

    /**
     * Creates a new {@link UnitsValue} from an existing one.
     *
     * @param other The {@link UnitsValue} to clone.
     */
    public UnitsValue(UnitsValue<T> other) {
        mValue = other.mValue;
        mUnits = other.mUnits;
    }

    /**
     * Creates a new {@link UnitsValue} from an existing one and converts it to the given {@link
     * Units}.
     *
     * @param other The {@link UnitsValue} to convert.
     * @param units The {@link Units} to use.
     */
    public UnitsValue(UnitsValue<T> other, T units) {
        mValue = units.convert(other.mUnits, other.mValue);
        mUnits = units;
    }

    /** @param other A {@link UnitsValue} to copy state from. */
    public void set(UnitsValue<T> other) {
        mValue = other.mValue;
        mUnits = other.mUnits;
    }

    /** @return The units. */
    public T getUnits() {
        return mUnits;
    }

    /** @param units The value to set for units. */
    public void setUnits(T units) {
        mUnits = units;
    }

    /** @return The value. */
    public Fixed6 getValue() {
        return mValue;
    }

    /** @param value The value to set for value. */
    public void setValue(Fixed6 value) {
        mValue = value;
    }

    /** @return The normalized value. */
    public Fixed6 getNormalizedValue() {
        return mUnits.normalize(mValue);
    }

    /** @param other The value to add to this one. */
    public void add(UnitsValue<?> other) {
        mValue = mValue.add(mUnits.convert(other.mUnits, other.mValue));
    }

    /** @param other The value to subtract from this one. */
    public void subtract(UnitsValue<?> other) {
        mValue = mValue.sub(mUnits.convert(other.mUnits, other.mValue));
    }

    /**
     * @return The default units to use during a load if nothing matches. {@code null} may be
     *         returned to indicate an error should occur instead.
     */
    public T getDefaultUnits() {
        return null;
    }

    @Override
    public int compareTo(UnitsValue<T> other) {
        if (this == other) {
            return 0;
        }
        return getNormalizedValue().compareTo(other.getNormalizedValue());
    }

    @Override
    public String toString() {
        return toString(true);
    }

    /**
     * @param localize Whether or not the number should be localized.
     * @return The formatted value.
     */
    public String toString(boolean localize) {
        return mUnits.format(mValue, localize);
    }

    @Override
    public boolean equals(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof UnitsValue<?>) {
            UnitsValue<?> uv = (UnitsValue<?>) obj;
            return mUnits == uv.mUnits && mValue.equals(uv.mValue);
        }
        return false;
    }

    @Override
    public int hashCode() {
        return super.hashCode();
    }
}
