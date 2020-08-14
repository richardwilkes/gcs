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

package com.trollworks.gcs.ui;

import java.awt.Color;
import java.util.ArrayList;
import java.util.List;

/** ThemeColor provides a dynamic color that tracks the current {@link Theme}. */
public class ThemeColor extends Color {
    public static final List<ThemeColor> ALL                                       = new ArrayList<>();
    public static final ThemeColor       BANDING                                   = new ThemeColor("banding");
    public static final ThemeColor       CURRENT_ENCUMBRANCE_BACKGROUND            = new ThemeColor("current_encumbrance_background");
    public static final ThemeColor       CURRENT_ENCUMBRANCE_BACKGROUND_OVERLOADED = new ThemeColor("current_encumbrance_background_overloaded");
    public static final ThemeColor       DIVIDER                                   = new ThemeColor("divider");
    public static final ThemeColor       SHEET_BACKGROUND                          = new ThemeColor("sheet_background");
    private final       int              mIndex;
    private final       String           mKey;

    private ThemeColor(String key) {
        super(0);
        mIndex = ALL.size();
        mKey = key;
        ALL.add(this);
    }

    /** @return The index to use for this {@link ThemeColor}. */
    public final int getIndex() {
        return mIndex;
    }

    /** @return The key to use for this {@link ThemeColor}. */
    public final String getKey() {
        return mKey;
    }

    @Override
    public final int getRGB() {
        return Theme.current().getColor(mIndex).getRGB();
    }

    @Override
    public final int hashCode() {
        return mIndex;
    }

    @Override
    public final boolean equals(Object obj) {
        return obj instanceof ThemeColor && ((ThemeColor) obj).mIndex == mIndex;
    }

    @Override
    public final String toString() {
        return Colors.encode(this);
    }
}
