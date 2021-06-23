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
    private final Color  mDefault;

    ThemeColor(String key, String name, Color def) {
        super(0, true);
        mName = name;
        mIndex = Colors.ALL.size();
        mKey = key;
        mDefault = new Color(def.getRGB(), true);
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

    /** @return The default color value. */
    public Color getDefault() {
        return mDefault;
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
        return obj instanceof ThemeColor && ((ThemeColor) obj).mIndex == mIndex;
    }

    @Override
    public String toString() {
        return mName;
    }
}
