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

    // The colors here intentionally avoid the pre-defined Color constants so that my IDE will
    // provide an interactive color swatch for letting me edit them inline.

    // Also note that the UI displays the colors in 4 evenly divided columns, in the order listed
    // in these definitions.

    public static final ThemeColor BACKGROUND             = new ThemeColor("background", I18n.text("Background"), new Color(238, 238, 238));
    public static final ThemeColor ON_BACKGROUND          = new ThemeColor("on_background", I18n.text("On Background"), new Color(0, 0, 0));
    public static final ThemeColor CONTENT                = new ThemeColor("content", I18n.text("Content"), new Color(255, 255, 255));
    public static final ThemeColor ON_CONTENT             = new ThemeColor("on_content", I18n.text("On Content"), new Color(0, 0, 0));
    public static final ThemeColor EDITABLE               = new ThemeColor("editable", I18n.text("Editable"), new Color(255, 255, 255));
    public static final ThemeColor ON_EDITABLE            = new ThemeColor("on_editable", I18n.text("On Editable"), new Color(0, 0, 160));
    public static final ThemeColor DISABLED_ON_EDITABLE   = new ThemeColor("disabled_on_editable", I18n.text("Disabled On Editable"), new Color(192, 192, 192));
    public static final ThemeColor EDITABLE_BORDER        = new ThemeColor("editable_border", I18n.text("Editable Border"), new Color(192, 192, 192));
    public static final ThemeColor ACTIVE_EDITABLE_BORDER = new ThemeColor("active_editable_border", I18n.text("Active Editable Border"), new Color(0, 0, 192));
    public static final ThemeColor BANDING                = new ThemeColor("banding", I18n.text("Banding"), new Color(235, 235, 220));
    public static final ThemeColor DIVIDER                = new ThemeColor("divider", I18n.text("Divider"), new Color(192, 192, 192));

    public static final ThemeColor HEADER     = new ThemeColor("header", I18n.text("Header"), new Color(43, 43, 43));
    public static final ThemeColor ON_HEADER  = new ThemeColor("on_header", I18n.text("On Header"), new Color(255, 255, 255));
    public static final ThemeColor MARKER     = new ThemeColor("marker", I18n.text("Marker"), new Color(252, 242, 196));
    public static final ThemeColor ON_CURRENT = new ThemeColor("on_marker", I18n.text("On Marker"), new Color(0, 0, 0));
    public static final ThemeColor WARNING    = new ThemeColor("warn", I18n.text("Warn"), new Color(255, 205, 210));
    public static final ThemeColor ON_WARNING = new ThemeColor("on_warn", I18n.text("On Warn"), new Color(0, 0, 0));
    public static final ThemeColor HINT       = new ThemeColor("hint", I18n.text("Hint"), new Color(128, 128, 128));

    public static final ThemeColor ACTIVE_TAB     = new ThemeColor("active_tab", I18n.text("Active Tab"), new Color(224, 212, 175));
    public static final ThemeColor ON_ACTIVE_TAB  = new ThemeColor("on_active_tab", I18n.text("On Active Tab"), new Color(0, 0, 0));
    public static final ThemeColor CURRENT_TAB    = new ThemeColor("current_tab", I18n.text("Current Tab"), new Color(211, 207, 197));
    public static final ThemeColor ON_CURRENT_TAB = new ThemeColor("on_current_tab", I18n.text("On Current Tab"), new Color(0, 0, 0));
    public static final ThemeColor PAGE           = new ThemeColor("page", I18n.text("Page"), new Color(255, 255, 255));
    public static final ThemeColor ON_PAGE        = new ThemeColor("on_page", I18n.text("On Page"), new Color(0, 0, 0));

    public static final ThemeColor BUTTON             = new ThemeColor("button.background", I18n.text("Button"), new Color(248, 248, 255));
    public static final ThemeColor ON_BUTTON          = new ThemeColor("on_button.background", I18n.text("On Button"), new Color(0, 0, 0));
    public static final ThemeColor ON_DISABLED_BUTTON = new ThemeColor("on_button.disabled", I18n.text("On Disabled Button"), new Color(192, 192, 192));
    public static final ThemeColor PRESSED_BUTTON     = new ThemeColor("button.pressed", I18n.text("Pressed Button"), new Color(0, 96, 160));
    public static final ThemeColor ON_PRESSED_BUTTON  = new ThemeColor("on_button.pressed", I18n.text("On Pressed Button"), new Color(255, 255, 255));
    public static final ThemeColor BUTTON_BORDER      = new ThemeColor("button.border", I18n.text("Button Border"), new Color(96, 96, 96));

    public static final ThemeColor ICON_BUTTON          = new ThemeColor("icon_button", I18n.text("Icon Button"), new Color(0, 0, 0));
    public static final ThemeColor DISABLED_ICON_BUTTON = new ThemeColor("disabled_icon_button", I18n.text("Disabled Icon Button"), new Color(192, 192, 192));
    public static final ThemeColor ROLLOVER_ICON_BUTTON = new ThemeColor("rollover_icon_button", I18n.text("Rollover Icon Button"), new Color(54, 137, 131));
    public static final ThemeColor PRESSED_ICON_BUTTON  = new ThemeColor("pressed_icon_button", I18n.text("Pressed Icon Button"), new Color(70, 171, 196));
    public static final ThemeColor DROP_AREA            = new ThemeColor("drop_area", I18n.text("Drop Area"), new Color(0, 0, 255));
    public static final ThemeColor PAGE_VOID            = new ThemeColor("page_void", I18n.text("Page Void"), new Color(192, 192, 192));

    public static final ThemeColor SELECTION    = new ThemeColor("selection", I18n.text("Selection"), new Color(0, 0, 224));
    public static final ThemeColor ON_SELECTION = new ThemeColor("on_selection", I18n.text("On Selection"), new Color(255, 255, 255));

    public static final ThemeColor SEARCH_LIST    = new ThemeColor("search_list", I18n.text("Search List"), new Color(224, 255, 255));
    public static final ThemeColor ON_SEARCH_LIST = new ThemeColor("on_search_list", I18n.text("On Search List"), new Color(0, 0, 0));

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
