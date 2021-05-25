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

import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.io.IOException;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.UIManager;

/** Provides standardized font access and utilities. */
public class Fonts {
    /** The name of the Font Awesome Brands font. */
    public static final  String             FONT_AWESOME_BRANDS  = "Font Awesome 5 Brands Regular";
    /** The name of the Font Awesome Regular font. */
    public static final  String             FONT_AWESOME_REGULAR = "Font Awesome 5 Free Regular";
    /** The name of the Font Awesome Solid font. */
    public static final  String             FONT_AWESOME_SOLID   = "Font Awesome 5 Free Solid";
    /** The name of the Roboto font. */
    public static final  String             ROBOTO               = "Roboto";
    /** The label font. */
    public static final  String             KEY_LABEL_PRIMARY    = "label.primary";
    /** The small label font. */
    public static final  String             KEY_LABEL_SECONDARY  = "label.secondary";
    /** The field font. */
    public static final  String             KEY_FIELD_PRIMARY    = "field.primary";
    /** The field notes font. */
    public static final  String             KEY_FIELD_SECONDARY  = "field.secondary";
    /** The primary footer font. */
    public static final  String             KEY_FOOTER_PRIMARY   = "footer.primary";
    /** The secondary footer font. */
    public static final  String             KEY_FOOTER_SECONDARY = "footer.secondary";
    private static final List<String>       KEYS                 = new ArrayList<>();
    private static final Map<String, Fonts> DEFAULTS             = new HashMap<>();
    private              String             mDescription;
    private              Font               mDefaultFont;

    private Fonts(String description, Font defaultFont) {
        mDescription = description;
        mDefaultFont = defaultFont;
    }

    /** Loads the current font settings from the preferences file. */
    public static void loadFromPreferences() {
        String[] embeddedFonts = {
                "Font Awesome 5 Brands-Regular-400.otf",
                "Font Awesome 5 Free-Regular-400.otf",
                "Font Awesome 5 Free-Solid-900.otf",
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
                "Roboto-ThinItalic.ttf"
        };
        for (String embeddedFont : embeddedFonts) {
            try (InputStream in = Fonts.class.getModule().getResourceAsStream("/fonts/" + embeddedFont)) {
                Font font = Font.createFont(Font.TRUETYPE_FONT, in);
                GraphicsEnvironment.getLocalGraphicsEnvironment().registerFont(font);
            } catch (Exception exception) {
                Log.error("unable to load font: " + embeddedFont);
            }
        }

        register(KEY_LABEL_PRIMARY, I18n.Text("Primary Labels"), new Font(ROBOTO, Font.PLAIN, 9));
        register(KEY_LABEL_SECONDARY, I18n.Text("Secondary Labels"), new Font(ROBOTO, Font.PLAIN, 8));
        register(KEY_FIELD_PRIMARY, I18n.Text("Primary Fields"), new Font(ROBOTO, Font.BOLD, 9));
        register(KEY_FIELD_SECONDARY, I18n.Text("Secondary Fields"), new Font(ROBOTO, Font.PLAIN, 8));
        register(KEY_FOOTER_PRIMARY, I18n.Text("Primary Footer"), new Font(ROBOTO, Font.BOLD, 8));
        register(KEY_FOOTER_SECONDARY, I18n.Text("Secondary Footer"), new Font(ROBOTO, Font.PLAIN, 6));
        Preferences prefs = Preferences.getInstance();
        for (String key : KEYS) {
            Info info = prefs.getFontInfo(key);
            if (info != null) {
                UIManager.put(key, info.create());
            }
        }
    }

    private static void register(String key, String description, Font defaultFont) {
        KEYS.add(key);
        UIManager.put(key, defaultFont);
        DEFAULTS.put(key, new Fonts(description, defaultFont));
    }

    /** Restores the default fonts. */
    public static void restoreDefaults() {
        Preferences prefs = Preferences.getInstance();
        for (String key : KEYS) {
            Font font = DEFAULTS.get(key).mDefaultFont;
            UIManager.put(key, font);
            prefs.setFontInfo(key, new Fonts.Info(font));
        }
    }

    /** @return Whether the fonts are currently at their default values or not. */
    public static boolean isSetToDefaults() {
        for (String key : KEYS) {
            if (!DEFAULTS.get(key).mDefaultFont.equals(UIManager.getFont(key))) {
                return false;
            }
        }
        return true;
    }

    /** @return The default system font to use. */
    public static Font getDefaultSystemFont() {
        return UIManager.getFont("TextField.font");
    }

    /** @return The available font keys. */
    public static String[] getKeys() {
        return KEYS.toArray(new String[0]);
    }

    /**
     * @param key The font key to lookup.
     * @return The human-readable label for the font.
     */
    public static String getDescription(String key) {
        Fonts match = DEFAULTS.get(key);
        return match != null ? match.mDescription : null;
    }

    /**
     * @param font The font to work on.
     * @return The font metrics for the specified font.
     */
    public static FontMetrics getFontMetrics(Font font) {
        Graphics2D  g2d = GraphicsUtilities.getGraphics();
        FontMetrics fm  = g2d.getFontMetrics(font);
        g2d.dispose();
        return fm;
    }

    public enum FontStyle {
        PLAIN {
            @Override
            public String toString() {
                return I18n.Text("Plain");
            }
        },
        BOLD {
            @Override
            public String toString() {
                return I18n.Text("Bold");
            }
        },
        ITALIC {
            @Override
            public String toString() {
                return I18n.Text("Italic");
            }
        },
        BOLD_ITALIC {
            @Override
            public String toString() {
                return I18n.Text("Bold Italic");
            }
        };

        public static FontStyle from(Font font) {
            Fonts.FontStyle[] styles = values();
            return styles[font.getStyle() % styles.length];
        }
    }

    public static class Info {
        private static final String    NAME  = "name";
        private static final String    STYLE = "style";
        private static final String    SIZE  = "size";
        public               String    mName;
        public               FontStyle mStyle;
        public               int       mSize;

        public Info(Font font) {
            mName = font.getName();
            mStyle = FontStyle.from(font);
            mSize = font.getSize();
        }

        public Info(JsonMap m) {
            mName = m.getStringWithDefault(NAME, ROBOTO);
            mStyle = Enums.extract(m.getString(STYLE), FontStyle.values(), FontStyle.PLAIN);
            mSize = m.getIntWithDefault(SIZE, 9);
            if (mSize < 1) {
                mSize = 1;
            }
        }

        public Font create() {
            return new Font(mName, mStyle.ordinal(), mSize);
        }

        public void toJSON(JsonWriter w) throws IOException {
            w.startMap();
            w.keyValue(NAME, mName);
            w.keyValue(STYLE, Enums.toId(mStyle));
            w.keyValue(SIZE, mSize);
            w.endMap();
        }
    }
}
