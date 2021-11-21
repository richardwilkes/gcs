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

package com.trollworks.gcs.ui;

import java.awt.Color;

/** ThemeColor provides a dynamic color that tracks the current {@link Colors}. */
public final class ThemeColor extends Color {
    private final int    mIndex;
    private final String mName;
    private final String mKey;
    private final Color  mDefLight;
    private final Color  mDefDark;

    ThemeColor(String key, String name, Color defLight, Color defDark) {
        super(0, true);
        mName = name;
        mIndex = Colors.ALL.size();
        mKey = key;
        mDefLight = new Color(defLight.getRGB(), true);
        mDefDark = new Color(defDark.getRGB(), true);
        Colors.ALL.add(this);
    }

    /** @return The index to use for this ThemeColor. */
    public int getIndex() {
        return mIndex;
    }

    /** @return The key to use for this ThemeColor. */
    public String getKey() {
        return mKey;
    }

    /** @return The default "light" color value. */
    public Color getDefaultLight() {
        return mDefLight;
    }

    /** @return The default "dark" color value. */
    public Color getDefaultDark() {
        return mDefDark;
    }

    @Override
    public int getRGB() {
        return Colors.currentThemeColors().getColor(mIndex).getRGB();
    }

    @Override
    public int hashCode() {
        return mIndex;
    }

    @Override
    public boolean equals(Object obj) {
        return obj instanceof ThemeColor tc && tc.mIndex == mIndex;
    }

    @Override
    public String toString() {
        return mName;
    }
}
