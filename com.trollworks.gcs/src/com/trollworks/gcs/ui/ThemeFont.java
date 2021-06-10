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

import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.List;

/**
 * ThemeFont allows the font to be changed dynamically. Unfortunately, there is no good way to
 * sub-class the Font class to do this, so we will rely on all code that wants to use this facility
 * to directly use ThemeFont rather than Font, which means not using the font accessible in the
 * Component object.
 */
public final class ThemeFont {
    /** The name of the Roboto font. */
    public static final String ROBOTO               = "Roboto";
    /** The name of the Roboto Black font. */
    public static final String ROBOTO_BLACK         = "Roboto Black";
    /** The name of the Font Awesome Brands font. */
    public static final String FONT_AWESOME_BRANDS  = "Font Awesome 5 Brands Regular";
    /** The name of the Font Awesome Regular font. */
    public static final String FONT_AWESOME_REGULAR = "Font Awesome 5 Free Regular";
    /** The name of the Font Awesome Solid font. */
    public static final String FONT_AWESOME_SOLID   = "Font Awesome 5 Free Solid";

    public static final List<ThemeFont> ALL = new ArrayList<>();

    public static final ThemeFont LABEL_PRIMARY;
    public static final ThemeFont FIELD_PRIMARY;
    public static final ThemeFont FIELD_SECONDARY;
    public static final ThemeFont PAGE_LABEL_PRIMARY;
    public static final ThemeFont PAGE_LABEL_SECONDARY;
    public static final ThemeFont PAGE_FIELD_PRIMARY;
    public static final ThemeFont PAGE_FIELD_SECONDARY;
    public static final ThemeFont PAGE_FOOTER_PRIMARY;
    public static final ThemeFont PAGE_FOOTER_SECONDARY;

    private final int    mIndex;
    private final String mName;
    private final String mKey;
    private final Font   mDefault;

    static {
        String[] embeddedFonts = {
                "Roboto-Black.ttf",
                "Roboto-BlackItalic.ttf",
                "Roboto-Bold.ttf",
                "Roboto-BoldItalic.ttf",
                "Roboto-Italic.ttf",
                "Roboto-Light.ttf",
                "Roboto-LightItalic.ttf",
                "Roboto-Medium.ttf",
                "Roboto-MediumItalic.ttf",
                "Roboto-Regular.ttf",
                "Roboto-Thin.ttf",
                "Roboto-ThinItalic.ttf",
                "Font Awesome 5 Brands-Regular-400.otf",
                "Font Awesome 5 Free-Regular-400.otf",
                "Font Awesome 5 Free-Solid-900.otf"
        };
        for (String embeddedFont : embeddedFonts) {
            try (InputStream in = Settings.class.getModule().getResourceAsStream("/fonts/" + embeddedFont)) {
                Font font = Font.createFont(Font.TRUETYPE_FONT, in);
                GraphicsEnvironment.getLocalGraphicsEnvironment().registerFont(font);
            } catch (Exception exception) {
                Log.error("unable to load font: " + embeddedFont);
            }
        }
        LABEL_PRIMARY = new ThemeFont("label.primary", I18n.text("Primary Labels"), new Font(ROBOTO, Font.PLAIN, 13));
        FIELD_PRIMARY = new ThemeFont("field.primary", I18n.text("Primary Fields"), new Font(ROBOTO_BLACK, Font.PLAIN, 13));
        FIELD_SECONDARY = new ThemeFont("field.secondary", I18n.text("Secondary Fields"), new Font(ROBOTO, Font.PLAIN, 11));
        PAGE_LABEL_PRIMARY = new ThemeFont("page.label.primary", I18n.text("Page Primary Labels"), new Font(ROBOTO, Font.PLAIN, 9));
        PAGE_LABEL_SECONDARY = new ThemeFont("page.label.secondary", I18n.text("Page Secondary Labels"), new Font(ROBOTO, Font.PLAIN, 8));
        PAGE_FIELD_PRIMARY = new ThemeFont("page.field.primary", I18n.text("Page Primary Fields"), new Font(ROBOTO_BLACK, Font.PLAIN, 9));
        PAGE_FIELD_SECONDARY = new ThemeFont("page.field.secondary", I18n.text("Page Secondary Fields"), new Font(ROBOTO, Font.PLAIN, 8));
        PAGE_FOOTER_PRIMARY = new ThemeFont("page.footer.primary", I18n.text("Page Primary Footer"), new Font(ROBOTO_BLACK, Font.PLAIN, 8));
        PAGE_FOOTER_SECONDARY = new ThemeFont("page.footer.secondary", I18n.text("Page Secondary Footer"), new Font(ROBOTO, Font.PLAIN, 6));
    }

    private ThemeFont(String key, String name, Font def) {
        mName = name;
        mIndex = ALL.size();
        mKey = key;
        mDefault = def;
        ALL.add(this);
    }

    /** @return The index to use for this ThemeFont. */
    public int getIndex() {
        return mIndex;
    }

    /** @return The key to use for this ThemeFont. */
    public String getKey() {
        return mKey;
    }

    /** @return The default font value. */
    public Font getDefault() {
        return mDefault;
    }

    /** @return The current font value. */
    public Font getFont() {
        return Theme.current().getFont(mIndex);
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
