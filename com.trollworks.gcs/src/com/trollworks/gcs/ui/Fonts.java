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

import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.StringTokenizer;
import javax.swing.UIManager;

/** Provides standardized font access and utilities. */
public class Fonts {
    /** The standard text field font. */
    public static final  String             KEY_STD_TEXT_FIELD    = "TextField.font";
    /** The label font. */
    public static final  String             KEY_LABEL_PRIMARY     = "label.primary";
    /** The small label font. */
    public static final  String             KEY_LABEL_SECONDARY   = "label.secondary";
    /** The field font. */
    public static final  String             KEY_FIELD_PRIMARY     = "field.primary";
    /** The field notes font. */
    public static final  String             KEY_FIELD_SECONDARY   = "field.secondary";
    /** The primary footer font. */
    public static final  String             KEY_FOOTER_PRIMARY    = "footer.primary";
    /** The secondary footer font. */
    public static final  String             KEY_FOOTER_SECONDARY  = "footer.secondary";
    /** The notification key used when font change notifications are broadcast. */
    public static final  String             FONT_NOTIFICATION_KEY = "FontsChanged";
    private static final List<String>       KEYS                  = new ArrayList<>();
    private static final Map<String, Fonts> DEFAULTS              = new HashMap<>();
    private              String             mDescription;
    private              Font               mDefaultFont;

    private Fonts(String description, Font defaultFont) {
        mDescription = description;
        mDefaultFont = defaultFont;
    }

    /** Loads the current font settings from the preferences file. */
    public static void loadFromPreferences() {
        String name = getDefaultFont().getName();
        register(KEY_LABEL_PRIMARY, I18n.Text("Primary Labels"), new Font(name, Font.PLAIN, 9));
        register(KEY_LABEL_SECONDARY, I18n.Text("Secondary Labels"), new Font(name, Font.PLAIN, 8));
        register(KEY_FIELD_PRIMARY, I18n.Text("Primary Fields"), new Font(name, Font.PLAIN, 9));
        register(KEY_FIELD_SECONDARY, I18n.Text("Secondary Fields"), new Font(name, Font.PLAIN, 8));
        register(KEY_FOOTER_PRIMARY, I18n.Text("Primary Footer"), new Font(name, Font.BOLD, 8));
        register(KEY_FOOTER_SECONDARY, I18n.Text("Secondary Footer"), new Font(name, Font.PLAIN, 6));
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

    /** @return The default font to use. */
    public static Font getDefaultFont() {
        return UIManager.getFont(KEY_STD_TEXT_FIELD);
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
     * @return The specified font as a canonical string.
     */
    public static String getStringValue(Font font) {
        return font.getName() + "," + font.getStyle() + "," + font.getSize();
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

    /**
     * @param buffer       The string to create the font from.
     * @param defaultValue The value to use if the string is invalid.
     * @return A font created from the specified string.
     */
    public static Font create(String buffer, Font defaultValue) {
        if (defaultValue == null) {
            defaultValue = getDefaultFont();
        }
        String name  = defaultValue.getName();
        int    style = defaultValue.getStyle();
        int    size  = defaultValue.getSize();
        if (buffer != null && !buffer.isEmpty()) {
            StringTokenizer tokenizer = new StringTokenizer(buffer, ",");
            if (tokenizer.hasMoreTokens()) {
                name = tokenizer.nextToken();
                if (!isValidFontName(name)) {
                    name = defaultValue.getName();
                }
                if (tokenizer.hasMoreTokens()) {
                    buffer = tokenizer.nextToken();
                    try {
                        style = Integer.parseInt(buffer);
                    } catch (NumberFormatException nfe1) {
                        // We'll use the default style instead
                    }
                    if (style < 0 || style > 3) {
                        style = defaultValue.getStyle();
                    }
                    if (tokenizer.hasMoreTokens()) {
                        buffer = tokenizer.nextToken();
                        try {
                            size = Integer.parseInt(buffer);
                        } catch (NumberFormatException nfe1) {
                            // We'll use the default size instead
                        }
                        if (size < 1) {
                            size = 1;
                        } else if (size > 200) {
                            size = 200;
                        }
                    }
                }
            }
        }
        return new Font(name, style, size);
    }

    /**
     * @param name The name to check.
     * @return {@code true} if the specified name is a valid font name.
     */
    public static boolean isValidFontName(String name) {
        for (String element : GraphicsEnvironment.getLocalGraphicsEnvironment().getAvailableFontFamilyNames()) {
            if (element.equalsIgnoreCase(name)) {
                return true;
            }
        }
        return false;
    }

    /** Cause font change listeners to be notified. */
    public static void notifyOfFontChanges() {
        Preferences.getInstance().getNotifier().notify(null, FONT_NOTIFICATION_KEY, null);
    }

    public enum FontStyle {
        PLAIN {
            @Override
            public String toString() {
                return I18n.Text("Plain");
            }
        }, BOLD {
            @Override
            public String toString() {
                return I18n.Text("Bold");
            }
        }, ITALIC {
            @Override
            public String toString() {
                return I18n.Text("Italic");
            }
        }, BOLD_ITALIC {
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
        private static final String    NAME   = "name";
        private static final String    STYLE  = "style";
        private static final String    SIZE   = "size";
        public static final  String[]  STYLES = {"plain", "bold", "italic", "Bold Italic"};
        public               String    mName;
        public               FontStyle mStyle;
        public               int       mSize;

        public Info(Font font) {
            mName = font.getName();
            mStyle = FontStyle.from(font);
            mSize = font.getSize();
        }

        public Info(JsonMap m) {
            mName = m.getStringWithDefault(NAME, "SansSerif");
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
