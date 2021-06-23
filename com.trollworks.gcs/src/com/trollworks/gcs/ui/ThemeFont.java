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

import java.awt.Font;

public class ThemeFont {
    interface Deriver {
        Font deriveFont();
    }

    private final int     mIndex;
    private final String  mName;
    private final String  mKey;
    private final Font    mDefault;
    private final Deriver mDeriver;

    ThemeFont(String key, String name, Font def) {
        mName = name;
        mIndex = Fonts.ALL.size();
        mKey = key;
        mDefault = def;
        mDeriver = null;
        Fonts.ALL.add(this);
    }

    ThemeFont(String key, Deriver deriver) {
        mName = key;
        mIndex = Fonts.ALL.size();
        mKey = key;
        mDeriver = deriver;
        mDefault = null;
        Fonts.ALL.add(this);
    }

    /** @return The index to use in the ALL list. */
    public int getIndex() {
        return mIndex;
    }

    /** @return The key to use in JSON. */
    public String getKey() {
        return mKey;
    }

    public boolean isEditable() {
        return mDeriver == null;
    }

    /** @return The default font value. */
    public Font getDefault() {
        return mDefault;
    }

    /** @return The current font value. */
    public Font getFont() {
        return mDeriver == null ? Fonts.currentThemeFonts().getFont(mIndex) : mDeriver.deriveFont();
    }

    @Override
    public int hashCode() {
        return mIndex;
    }

    @Override
    public boolean equals(Object obj) {
        return obj instanceof ThemeFont && ((ThemeFont) obj).mIndex == mIndex;
    }

    @Override
    public String toString() {
        return mName;
    }
}
