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

import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.util.ArrayList;
import java.util.List;

/** ThemeColor provides a dynamic color that tracks the current {@link Theme}. */
public final class ThemeColor extends Color {
    public static final List<ThemeColor> ALL = new ArrayList<>();

    public static final ThemeColor CURRENT    = new ThemeColor("current", I18n.text("Current"), new Color(252, 242, 196));
    public static final ThemeColor ON_CURRENT = new ThemeColor("on_current", I18n.text("On Current"), Color.BLACK);
    public static final ThemeColor WARN       = new ThemeColor("warn", I18n.text("Warn"), new Color(255, 205, 210));
    public static final ThemeColor ON_WARN    = new ThemeColor("on_warn", I18n.text("On Warn"), Color.BLACK);

    public static final ThemeColor PAGE      = new ThemeColor("page", I18n.text("Page"), Color.WHITE);
    public static final ThemeColor ON_PAGE   = new ThemeColor("on_page", I18n.text("On Page"), Color.BLACK);
    public static final ThemeColor PAGE_VOID = new ThemeColor("page_void", I18n.text("Page Void"), Color.LIGHT_GRAY);
    public static final ThemeColor DIVIDER   = new ThemeColor("divider", I18n.text("Divider"), Color.LIGHT_GRAY);

    public static final ThemeColor HEADER     = new ThemeColor("header", I18n.text("Header"), new Color(43, 43, 43));
    public static final ThemeColor ON_HEADER  = new ThemeColor("on_header", I18n.text("On Header"), Color.WHITE);
    public static final ThemeColor CONTENT    = new ThemeColor("content", I18n.text("Content"), new Color(220, 220, 210));
    public static final ThemeColor ON_CONTENT = new ThemeColor("on_content", I18n.text("On Content"), Color.BLACK);

    public static final ThemeColor EDITABLE_LINE = new ThemeColor("editable_line", I18n.text("Editable Line"), Color.LIGHT_GRAY);
    public static final ThemeColor ON_EDITABLE   = new ThemeColor("on_editable", I18n.text("On Editable"), new Color(0, 0, 160));
    public static final ThemeColor BANDING       = new ThemeColor("banding", I18n.text("Banding"), new Color(200, 200, 185));

    private final int    mIndex;
    private final String mName;
    private final String mKey;
    private final Color  mDefault;

    private ThemeColor(String key, String name, Color def) {
        super(0, true);
        mName = name;
        mIndex = ALL.size();
        mKey = key;
        mDefault = new Color(def.getRGB(), true);
        ALL.add(this);
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
        return Theme.current().getColor(mIndex).getRGB();
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
