/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui;

import com.trollworks.gcs.utility.Preferences;

import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.GraphicsEnvironment;
import java.util.Map.Entry;
import java.util.StringTokenizer;
import java.util.TreeMap;
import javax.swing.UIManager;

/** Provides standardized font access and utilities. */
public class Fonts {
    /** The notification key used when font change notifications are broadcast. */
    public static final  String                 FONT_NOTIFICATION_KEY = "FontsChanged";
    private static final String                 MODULE                = "Font";
    /** The standard text field font. */
    public static final  String                 KEY_STD_TEXT_FIELD    = "TextField.font";
    private static final TreeMap<String, Fonts> DEFAULTS              = new TreeMap<>();
    private              String                 mDescription;
    private              Font                   mDefaultFont;

    /**
     * Registers a default for a specific font key.
     *
     * @param key         The key the font maps to.
     * @param description A human-readable label for the font.
     * @param defaultFont The default font.
     */
    public static void register(String key, String description, Font defaultFont) {
        UIManager.put(key, defaultFont);
        DEFAULTS.put(key, new Fonts(description, defaultFont));
    }

    /** @return The available font keys. */
    public static String[] getKeys() {
        return DEFAULTS.keySet().toArray(new String[0]);
    }

    /**
     * @param key The font key to lookup.
     * @return The human-readable label for the font.
     */
    public static String getDescription(String key) {
        Fonts match = DEFAULTS.get(key);
        return match != null ? match.mDescription : null;
    }

    /** @return The default font to use. */
    public static Font getDefaultFont() {
        return UIManager.getFont(KEY_STD_TEXT_FIELD);
    }

    /** @return The default font name to use. */
    public static String getDefaultFontName() {
        return getDefaultFont().getName();
    }

    /** Restores the default fonts. */
    public static void restoreDefaults() {
        for (Entry<String, Fonts> entry : DEFAULTS.entrySet()) {
            UIManager.put(entry.getKey(), entry.getValue().mDefaultFont);
        }
    }

    /** @return Whether the fonts are currently at their default values or not. */
    public static boolean isSetToDefaults() {
        for (Entry<String, Fonts> entry : DEFAULTS.entrySet()) {
            if (!entry.getValue().mDefaultFont.equals(UIManager.getFont(entry.getKey()))) {
                return false;
            }
        }
        return true;
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
     * @param buffer The string to create the font from.
     * @return A font created from the specified string.
     */
    public static Font create(String buffer) {
        return create(buffer, null);
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

    /** Loads the current font settings from the preferences file. */
    public static void loadFromPreferences() {
        Preferences prefs = Preferences.getInstance();
        for (String key : DEFAULTS.keySet()) {
            Font font = prefs.getFontValue(MODULE, key);
            if (font != null) {
                UIManager.put(key, font);
            }
        }
    }

    /** Saves the current font settings to the preferences file. */
    public static void saveToPreferences() {
        Preferences prefs = Preferences.getInstance();
        prefs.removePreferences(MODULE);
        for (String key : DEFAULTS.keySet()) {
            Font font = UIManager.getFont(key);
            if (font != null) {
                prefs.setValue(MODULE, key, font);
            }
        }
    }

    /** Cause font change listeners to be notified. */
    public static void notifyOfFontChanges() {
        Preferences.getInstance().getNotifier().notify(null, FONT_NOTIFICATION_KEY, null);
    }

    private Fonts(String description, Font defaultFont) {
        mDescription = description;
        mDefaultFont = defaultFont;
    }
}
