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
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Font;
import java.awt.GraphicsEnvironment;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

/**
 * Fonts allows the font to be changed dynamically. Unfortunately, there is no good way to sub-class
 * the Font class to do this, so we will rely on all code that wants to use this facility to
 * directly use Fonts rather than Font, which means not using the font accessible in the Component
 * object.
 */
public final class Fonts {
    /** The name of the Roboto font. */
    public static final String ROBOTO               = "Roboto";
    /** The name of the Roboto Black font. */
    public static final String ROBOTO_BLACK         = "Roboto Black";
    /** The name of the Roboto Medium font. */
    public static final String ROBOTO_MEDIUM        = "Roboto Medium";
    /** The name of the Font Awesome Brands font. */
    public static final String FONT_AWESOME_BRANDS  = "Font Awesome 5 Brands Regular";
    /** The name of the Font Awesome Regular font. */
    public static final String FONT_AWESOME_REGULAR = "Font Awesome 5 Free Regular";
    /** The name of the Font Awesome Solid font. */
    public static final String FONT_AWESOME_SOLID   = "Font Awesome 5 Free Solid";
    /** The name of the RPG Awesome font. */
    public static final String RPG_AWESOME          = "rpg-awesome";

    public static final  List<ThemeFont> ALL             = new ArrayList<>();
    private static final int             MINIMUM_VERSION = 1;
    private static final int             CURRENT_VERSION = 1;

    private static final Font  FALLBACK_FONT = new Font(Font.DIALOG, Font.PLAIN, 12);
    private static final Fonts DEFAULTS;
    private static       Fonts CURRENT;

    public static final ThemeFont BUTTON;
    public static final ThemeFont HEADER;
    public static final ThemeFont LABEL_PRIMARY;
    public static final ThemeFont LABEL_SECONDARY;
    public static final ThemeFont FIELD_PRIMARY;
    public static final ThemeFont FIELD_SECONDARY;
    public static final ThemeFont TOOLTIP;
    public static final ThemeFont PAGE_LABEL_PRIMARY;
    public static final ThemeFont PAGE_LABEL_SECONDARY;
    public static final ThemeFont PAGE_FIELD_PRIMARY;
    public static final ThemeFont PAGE_FIELD_SECONDARY;
    public static final ThemeFont PAGE_FOOTER_PRIMARY;
    public static final ThemeFont PAGE_FOOTER_SECONDARY;

    // Derived theme fonts
    public static final ThemeFont KEYBOARD;
    public static final ThemeFont ENCUMBRANCE_MARKER;
    public static final ThemeFont FONT_ICON_STD;
    public static final ThemeFont FONT_ICON_LABEL_PRIMARY;
    public static final ThemeFont FONT_ICON_PAGE_SMALL;
    public static final ThemeFont FONT_ICON_HUGE;
    public static final ThemeFont FONT_ICON_FILE_RPG;
    public static final ThemeFont FONT_ICON_FILE_FA;

    private           Font[]  mFonts;
    private transient boolean mReadOnly;

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
                "Font Awesome 5 Free-Solid-900.otf",
                "RPG Awesome Webfont.ttf"
        };
        for (String embeddedFont : embeddedFonts) {
            try (InputStream in = Settings.class.getModule().getResourceAsStream("/fonts/" + embeddedFont)) {
                Font font = Font.createFont(Font.TRUETYPE_FONT, in);
                GraphicsEnvironment.getLocalGraphicsEnvironment().registerFont(font);
            } catch (Exception exception) {
                Log.error("unable to load font: " + embeddedFont);
            }
        }
        BUTTON = new ThemeFont("button", I18n.text("Button"), new Font(ROBOTO_BLACK, Font.PLAIN, 13));
        HEADER = new ThemeFont("header", I18n.text("Header"), new Font(ROBOTO_MEDIUM, Font.PLAIN, 13));
        LABEL_PRIMARY = new ThemeFont("label.primary", I18n.text("Primary Labels"), new Font(ROBOTO, Font.PLAIN, 13));
        LABEL_SECONDARY = new ThemeFont("label.secondary", I18n.text("Secondary Labels"), new Font(ROBOTO, Font.PLAIN, 11));
        FIELD_PRIMARY = new ThemeFont("field.primary", I18n.text("Primary Fields"), new Font(ROBOTO, Font.PLAIN, 13));
        FIELD_SECONDARY = new ThemeFont("field.secondary", I18n.text("Secondary Fields"), new Font(ROBOTO, Font.PLAIN, 11));
        TOOLTIP = new ThemeFont("tooltip", I18n.text("Tooltip"), new Font(ROBOTO, Font.PLAIN, 12));
        PAGE_LABEL_PRIMARY = new ThemeFont("page.label.primary", I18n.text("Page Primary Labels"), new Font(ROBOTO, Font.PLAIN, 9));
        PAGE_LABEL_SECONDARY = new ThemeFont("page.label.secondary", I18n.text("Page Secondary Labels"), new Font(ROBOTO, Font.PLAIN, 8));
        PAGE_FIELD_PRIMARY = new ThemeFont("page.field.primary", I18n.text("Page Primary Fields"), new Font(ROBOTO_MEDIUM, Font.PLAIN, 9));
        PAGE_FIELD_SECONDARY = new ThemeFont("page.field.secondary", I18n.text("Page Secondary Fields"), new Font(ROBOTO, Font.PLAIN, 8));
        PAGE_FOOTER_PRIMARY = new ThemeFont("page.footer.primary", I18n.text("Page Primary Footer"), new Font(ROBOTO_MEDIUM, Font.PLAIN, 8));
        PAGE_FOOTER_SECONDARY = new ThemeFont("page.footer.secondary", I18n.text("Page Secondary Footer"), new Font(ROBOTO, Font.PLAIN, 6));

        KEYBOARD = new ThemeFont("keyboard", () -> new Font("Dialog", Font.PLAIN, LABEL_PRIMARY.getFont().getSize()));
        ENCUMBRANCE_MARKER = new ThemeFont("encumbrance.marker", () -> new Font(FONT_AWESOME_SOLID, Font.PLAIN, PAGE_LABEL_PRIMARY.getFont().getSize()));
        FONT_ICON_STD = new ThemeFont("fonticon.std", () -> new Font(FONT_AWESOME_SOLID, Font.PLAIN, BUTTON.getFont().getSize()));
        FONT_ICON_LABEL_PRIMARY = new ThemeFont("fonticon.label.primary", () -> new Font(FONT_AWESOME_SOLID, Font.PLAIN, LABEL_PRIMARY.getFont().getSize()));
        FONT_ICON_PAGE_SMALL = new ThemeFont("fonticon.page.small", () -> new Font(FONT_AWESOME_SOLID, Font.PLAIN, PAGE_LABEL_SECONDARY.getFont().getSize()));
        FONT_ICON_HUGE = new ThemeFont("fonticon.huge", () -> new Font(FONT_AWESOME_SOLID, Font.PLAIN, LABEL_PRIMARY.getFont().getSize() * 3));
        FONT_ICON_FILE_RPG = new ThemeFont("fonticon.file.rpg", () -> new Font(RPG_AWESOME, Font.PLAIN, LABEL_PRIMARY.getFont().getSize()));
        FONT_ICON_FILE_FA = new ThemeFont("fonticon.file.fa", () -> new Font(FONT_AWESOME_SOLID, Font.PLAIN, LABEL_PRIMARY.getFont().getSize()));

        DEFAULTS = new Fonts();
        for (ThemeFont font : ALL) {
            DEFAULTS.setFont(font.getIndex(), font.getDefault());
        }
        DEFAULTS.mReadOnly = true;
        CURRENT = new Fonts(DEFAULTS);
    }

    /** @return The current theme fonts. */
    public static Fonts defaultThemeFonts() {
        return DEFAULTS;
    }

    /** @return The current theme fonts. */
    public static Fonts currentThemeFonts() {
        return CURRENT;
    }

    /**
     * Set the current theme fonts.
     *
     * @param fonts The theme fonts to set as current.
     */
    public static void setCurrentThemeFonts(Fonts fonts) {
        CURRENT = new Fonts(fonts);
    }

    private Fonts() {
        mFonts = new Font[ALL.size()];
        for (int i = mFonts.length - 1; i >= 0; i--) {
            mFonts[i] = FALLBACK_FONT;
        }
    }

    /**
     * Creates theme fonts from an existing set.
     *
     * @param other The other theme fonts to base this one off of.
     */
    public Fonts(Fonts other) {
        mFonts = new Font[ALL.size()];
        System.arraycopy(other.mFonts, 0, mFonts, 0, mFonts.length);
    }

    /**
     * Creates theme fonts from a file.
     *
     * @param path The path to load the theme fonts from.
     */
    public Fonts(Path path) throws IOException {
        this(DEFAULTS);
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m       = Json.asMap(Json.parse(in));
            int     version = m.getInt(Settings.VERSION);
            if (version >= MINIMUM_VERSION && version <= CURRENT_VERSION && m.has(Settings.FONTS)) {
                load(m.getMap(Settings.FONTS));
            }
        }
    }

    /**
     * Creates theme fonts from a JsonMap.
     *
     * @param m The map to load the theme fonts from.
     */
    public Fonts(JsonMap m) {
        this(DEFAULTS);
        load(m);
    }

    private void load(JsonMap m) {
        for (ThemeFont one : ALL) {
            if (one.isEditable() && m.has(one.getKey())) {
                setFont(one.getIndex(), new FontDesc(m.getMap(one.getKey())).create());
            }
        }
    }

    /**
     * Save the theme fonts to a file.
     *
     * @param path The path to write to.
     */
    public void save(Path path) throws IOException {
        SafeFileUpdater trans = new SafeFileUpdater();
        trans.begin();
        try {
            Files.createDirectories(path.getParent());
            File file = trans.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                w.startMap();
                w.keyValue(Settings.VERSION, CURRENT_VERSION);
                w.key(Settings.FONTS);
                save(w);
                w.endMap();
            }
        } catch (IOException ioe) {
            trans.abort();
            throw ioe;
        }
        trans.commit();
    }

    /**
     * Save the theme fonts to a JsonWriter.
     *
     * @param w The JsonWriter to write to.
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        for (ThemeFont one : ALL) {
            if (one.isEditable()) {
                w.key(one.getKey());
                new FontDesc(getFont(one.getIndex())).toJSON(w);
            }
        }
        w.endMap();
    }

    /** @return {@code true} if this a stock theme and is therefore unmodifiable. */
    public boolean isReadOnly() {
        return mReadOnly;
    }

    /**
     * @param index The index to retrieve the {@link Font} for.
     * @return The {@link Font} for the index.
     */
    public Font getFont(int index) {
        if (index < 0 || index >= mFonts.length) {
            return FALLBACK_FONT;
        }
        return mFonts[index];
    }

    /**
     * @param index The index to set the {@link Font} for.
     * @param font  The new {@link Font} to use.
     */
    public void setFont(int index, Font font) {
        if (!mReadOnly && index >= 0 && index < mFonts.length) {
            mFonts[index] = font;
        }
    }
}
