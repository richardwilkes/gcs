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
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Color;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** Provides standardized color access. */
public final class Colors {
    public static final  List<ThemeColor>    ALL             = new ArrayList<>();
    public static final  Color               TRANSPARENT     = new Color(0, 0, 0, 0);
    private static final Map<String, String> NAME_TO_RGB     = new HashMap<>();
    private static final Map<String, String> RGB_TO_NAME     = new HashMap<>();
    private static final int                 MINIMUM_VERSION = 1;
    private static final int                 CURRENT_VERSION = 1;

    // The theme colors here intentionally avoid the pre-defined Color constants so that my IDE will
    // provide an interactive color swatch for letting me edit them inline.

    // Also note that the color settings UI displays the colors in 4 evenly divided columns, in the
    // order listed in these definitions.

    public static final ThemeColor BACKGROUND     = new ThemeColor("background", I18n.text("Background"),
            new Color(238, 238, 238),
            new Color(50, 50, 50));
    public static final ThemeColor ON_BACKGROUND  = new ThemeColor("on_background", I18n.text("On Background"),
            new Color(0, 0, 0),
            new Color(221, 221, 221));
    public static final ThemeColor CONTENT        = new ThemeColor("content", I18n.text("Content"),
            new Color(255, 255, 255),
            new Color(32, 32, 32));
    public static final ThemeColor ON_CONTENT     = new ThemeColor("on_content", I18n.text("On Content"),
            new Color(0, 0, 0),
            new Color(221, 221, 221));
    public static final ThemeColor BANDING        = new ThemeColor("banding", I18n.text("Banding"),
            new Color(235, 235, 220),
            new Color(42, 42, 42));
    public static final ThemeColor DIVIDER        = new ThemeColor("divider", I18n.text("Divider"),
            new Color(192, 192, 192),
            new Color(102, 102, 102));
    public static final ThemeColor HEADER         = new ThemeColor("header", I18n.text("Header"),
            new Color(43, 43, 43),
            new Color(64, 64, 64));
    public static final ThemeColor ON_HEADER      = new ThemeColor("on_header", I18n.text("On Header"),
            new Color(255, 255, 255),
            new Color(192, 192, 192));
    public static final ThemeColor TAB_FOCUSED    = new ThemeColor("tab_focused", I18n.text("Focused Tab"),
            new Color(224, 212, 175),
            new Color(102, 102, 0));
    public static final ThemeColor ON_TAB_FOCUSED = new ThemeColor("on_tab_focused", I18n.text("On Focused Tab"),
            new Color(0, 0, 0),
            new Color(221, 221, 221));
    public static final ThemeColor TAB_CURRENT    = new ThemeColor("tab_current", I18n.text("Current Tab"),
            new Color(211, 207, 197),
            new Color(61, 61, 0));
    public static final ThemeColor ON_TAB_CURRENT = new ThemeColor("on_tab_current", I18n.text("On Current Tab"),
            new Color(0, 0, 0),
            new Color(221, 221, 221));
    public static final ThemeColor DROP_AREA      = new ThemeColor("drop_area", I18n.text("Drop Area"),
            new Color(204, 0, 51),
            new Color(255, 0, 0));

    public static final ThemeColor EDITABLE                = new ThemeColor("editable", I18n.text("Editable"),
            new Color(255, 255, 255),
            new Color(24, 24, 24));
    public static final ThemeColor ON_EDITABLE             = new ThemeColor("on_editable", I18n.text("On Editable"),
            new Color(0, 0, 160),
            new Color(0, 153, 153));
    public static final ThemeColor EDITABLE_BORDER         = new ThemeColor("editable_border", I18n.text("Editable Border"),
            new Color(192, 192, 192),
            new Color(96, 96, 96));
    public static final ThemeColor EDITABLE_BORDER_FOCUSED = new ThemeColor("editable_border_focused", I18n.text("Focused Editable Border"),
            new Color(0, 0, 192),
            new Color(0, 102, 102));
    public static final ThemeColor SELECTION               = new ThemeColor("selection", I18n.text("Selection"),
            new Color(0, 96, 160),
            new Color(0, 96, 160));
    public static final ThemeColor ON_SELECTION            = new ThemeColor("on_selection", I18n.text("On Selection"),
            new Color(255, 255, 255),
            new Color(255, 255, 255));
    public static final ThemeColor INACTIVE_SELECTION      = new ThemeColor("inactive_selection", I18n.text("Inactive Selection"),
            new Color(0, 64, 148),
            new Color(0, 64, 148));
    public static final ThemeColor ON_INACTIVE_SELECTION   = new ThemeColor("on_inactive_selection", I18n.text("On Inactive Selection"),
            new Color(228, 228, 228),
            new Color(228, 228, 228));
    public static final ThemeColor SCROLL                  = new ThemeColor("scroll", I18n.text("Scroll"),
            new Color(192, 192, 192, 128),
            new Color(128, 128, 128, 128));
    public static final ThemeColor SCROLL_ROLLOVER         = new ThemeColor("scroll_rollover", I18n.text("Scroll Rollover"),
            new Color(192, 192, 192),
            new Color(128, 128, 128));
    public static final ThemeColor SCROLL_EDGE             = new ThemeColor("scroll_edge", I18n.text("Scroll Edge"),
            new Color(128, 128, 128),
            new Color(160, 160, 160));
    public static final ThemeColor ACCENT                  = new ThemeColor("accent", I18n.text("Accent"),
            new Color(0, 102, 102),
            new Color(0, 153, 153));

    public static final ThemeColor CONTROL              = new ThemeColor("control", I18n.text("Control"),
            new Color(248, 248, 255),
            new Color(64, 64, 64));
    public static final ThemeColor ON_CONTROL           = new ThemeColor("on_control", I18n.text("On Control"),
            new Color(0, 0, 0),
            new Color(221, 221, 221));
    public static final ThemeColor CONTROL_PRESSED      = new ThemeColor("control_pressed", I18n.text("Pressed Control"),
            new Color(0, 96, 160),
            new Color(0, 96, 160));
    public static final ThemeColor ON_CONTROL_PRESSED   = new ThemeColor("on_control_pressed", I18n.text("On Pressed Control"),
            new Color(255, 255, 255),
            new Color(255, 255, 255));
    public static final ThemeColor CONTROL_EDGE         = new ThemeColor("control_edge", I18n.text("Control Edge"),
            new Color(96, 96, 96),
            new Color(96, 96, 96));
    public static final ThemeColor ICON_BUTTON          = new ThemeColor("icon_button", I18n.text("Icon Button"),
            new Color(96, 96, 96),
            new Color(128, 128, 128));
    public static final ThemeColor ICON_BUTTON_ROLLOVER = new ThemeColor("icon_button_rollover", I18n.text("Icon Button Rollover"),
            new Color(0, 0, 0),
            new Color(192, 192, 192));
    public static final ThemeColor ICON_BUTTON_PRESSED  = new ThemeColor("icon_button_pressed", I18n.text("Pressed Icon Button"),
            new Color(0, 96, 160),
            new Color(0, 96, 160));
    public static final ThemeColor TOOLTIP              = new ThemeColor("tooltip", I18n.text("Tooltip"),
            new Color(252, 252, 196),
            new Color(153, 153, 0));
    public static final ThemeColor ON_TOOLTIP           = new ThemeColor("on_tooltip", I18n.text("On Tooltip"),
            new Color(0, 0, 0),
            new Color(32, 32, 32));
    public static final ThemeColor SEARCH_LIST          = new ThemeColor("search_list", I18n.text("Search List"),
            new Color(224, 255, 255),
            new Color(0, 43, 43));
    public static final ThemeColor ON_SEARCH_LIST       = new ThemeColor("on_search_list", I18n.text("On Search List"),
            new Color(0, 0, 0),
            new Color(204, 204, 204));

    public static final ThemeColor PAGE          = new ThemeColor("page", I18n.text("Page"),
            new Color(255, 255, 255),
            new Color(16, 16, 16));
    public static final ThemeColor ON_PAGE       = new ThemeColor("on_page", I18n.text("On Page"),
            new Color(0, 0, 0),
            new Color(160, 160, 160));
    public static final ThemeColor PAGE_VOID     = new ThemeColor("page_void", I18n.text("Page Void"),
            new Color(128, 128, 128),
            new Color(0, 0, 0));
    public static final ThemeColor MARKER        = new ThemeColor("marker", I18n.text("Marker"),
            new Color(252, 242, 196),
            new Color(0, 51, 0));
    public static final ThemeColor ON_MARKER     = new ThemeColor("on_marker", I18n.text("On Marker"),
            new Color(0, 0, 0),
            new Color(221, 221, 221));
    public static final ThemeColor ERROR         = new ThemeColor("error", I18n.text("Error"),
            new Color(128, 0, 0),
            new Color(128, 0, 0));
    public static final ThemeColor ON_ERROR      = new ThemeColor("on_error", I18n.text("On Error"),
            new Color(255, 255, 255),
            new Color(221, 221, 221));
    public static final ThemeColor WARNING       = new ThemeColor("warning", I18n.text("Warning"),
            new Color(128, 64, 0),
            new Color(153, 102, 0));
    public static final ThemeColor ON_WARNING    = new ThemeColor("on_warning", I18n.text("On Warning"),
            new Color(255, 255, 255),
            new Color(221, 221, 221));
    public static final ThemeColor OVERLOADED    = new ThemeColor("overloaded", I18n.text("Overloaded"),
            new Color(192, 64, 64),
            new Color(115, 37, 37));
    public static final ThemeColor ON_OVERLOADED = new ThemeColor("on_overloaded", I18n.text("On Overloaded"),
            new Color(255, 255, 255),
            new Color(221, 221, 221));
    public static final ThemeColor HINT          = new ThemeColor("hint", I18n.text("Hint"),
            new Color(128, 128, 128),
            new Color(64, 64, 64));

    private static final Colors LIGHT_DEFAULTS;
    private static final Colors DARK_DEFAULTS;
    private static       Colors CURRENT;

    private           Color[] mColors;
    private transient boolean mReadOnly;

    static {
        // The HTML / CSS color list
        add("AliceBlue", "240,248,255");
        add("AntiqueWhite", "250,235,215");
        add("Aqua", "0,255,255");
        add("Aquamarine", "127,255,212");
        add("Azure", "240,255,255");
        add("Beige", "245,245,220");
        add("Bisque", "255,228,196");
        add("Black", "0,0,0");
        add("BlanchedAlmond", "255,235,205");
        add("Blue", "0,0,255");
        add("BlueViolet", "138,43,226");
        add("Brown", "165,42,42");
        add("BurlyWood", "222,184,135");
        add("CadetBlue", "95,158,160");
        add("Chartreuse", "127,255,0");
        add("Chocolate", "210,105,30");
        add("Coral", "255,127,80");
        add("CornflowerBlue", "100,149,237");
        add("Cornsilk", "255,248,220");
        add("Crimson", "220,20,60");
        add("Cyan", "0,255,255");
        add("DarkBlue", "0,0,139");
        add("DarkCyan", "0,139,139");
        add("DarkGoldenRod", "184,134,11");
        add("DarkGray", "169,169,169");
        add("DarkGrey", "169,169,169");
        add("DarkGreen", "0,100,0");
        add("DarkKhaki", "189,183,107");
        add("DarkMagenta", "139,0,139");
        add("DarkOliveGreen", "85,107,47");
        add("DarkOrange", "255,140,0");
        add("DarkOrchid", "153,50,204");
        add("DarkRed", "139,0,0");
        add("DarkSalmon", "233,150,122");
        add("DarkSeaGreen", "143,188,143");
        add("DarkSlateBlue", "72,61,139");
        add("DarkSlateGray", "47,79,79");
        add("DarkSlateGrey", "47,79,79");
        add("DarkTurquoise", "0,206,209");
        add("DarkViolet", "148,0,211");
        add("DeepPink", "255,20,147");
        add("DeepSkyBlue", "0,191,255");
        add("DimGray", "105,105,105");
        add("DimGrey", "105,105,105");
        add("DodgerBlue", "30,144,255");
        add("FireBrick", "178,34,34");
        add("FloralWhite", "255,250,240");
        add("ForestGreen", "34,139,34");
        add("Fuchsia", "255,0,255");
        add("Gainsboro", "220,220,220");
        add("GhostWhite", "248,248,255");
        add("Gold", "255,215,0");
        add("GoldenRod", "218,165,32");
        add("Gray", "128,128,128");
        add("Grey", "128,128,128");
        add("Green", "0,128,0");
        add("GreenYellow", "173,255,47");
        add("HoneyDew", "240,255,240");
        add("HotPink", "255,105,180");
        add("IndianRed", "205,92,92");
        add("Indigo", "75,0,130");
        add("Ivory", "255,255,240");
        add("Khaki", "240,230,140");
        add("Lavender", "230,230,250");
        add("LavenderBlush", "255,240,245");
        add("LawnGreen", "124,252,0");
        add("LemonChiffon", "255,250,205");
        add("LightBlue", "173,216,230");
        add("LightCoral", "240,128,128");
        add("LightCyan", "224,255,255");
        add("LightGoldenRodYellow", "250,250,210");
        add("LightGray", "211,211,211");
        add("LightGrey", "211,211,211");
        add("LightGreen", "144,238,144");
        add("LightPink", "255,182,193");
        add("LightSalmon", "255,160,122");
        add("LightSeaGreen", "32,178,170");
        add("LightSkyBlue", "135,206,250");
        add("LightSlateGray", "119,136,153");
        add("LightSlateGrey", "119,136,153");
        add("LightSteelBlue", "176,196,222");
        add("LightYellow", "255,255,224");
        add("Lime", "0,255,0");
        add("LimeGreen", "50,205,50");
        add("Linen", "250,240,230");
        add("Magenta", "255,0,255");
        add("Maroon", "128,0,0");
        add("MediumAquaMarine", "102,205,170");
        add("MediumBlue", "0,0,205");
        add("MediumOrchid", "186,85,211");
        add("MediumPurple", "147,112,219");
        add("MediumSeaGreen", "60,179,113");
        add("MediumSlateBlue", "123,104,238");
        add("MediumSpringGreen", "0,250,154");
        add("MediumTurquoise", "72,209,204");
        add("MediumVioletRed", "199,21,133");
        add("MidnightBlue", "25,25,112");
        add("MintCream", "245,255,250");
        add("MistyRose", "255,228,225");
        add("Moccasin", "255,228,181");
        add("NavajoWhite", "255,222,173");
        add("Navy", "0,0,128");
        add("OldLace", "253,245,230");
        add("Olive", "128,128,0");
        add("OliveDrab", "107,142,35");
        add("Orange", "255,165,0");
        add("OrangeRed", "255,69,0");
        add("Orchid", "218,112,214");
        add("PaleGoldenRod", "238,232,170");
        add("PaleGreen", "152,251,152");
        add("PaleTurquoise", "175,238,238");
        add("PaleVioletRed", "219,112,147");
        add("PapayaWhip", "255,239,213");
        add("PeachPuff", "255,218,185");
        add("Peru", "205,133,63");
        add("Pink", "255,192,203");
        add("Plum", "221,160,221");
        add("PowderBlue", "176,224,230");
        add("Purple", "128,0,128");
        add("Red", "255,0,0");
        add("RosyBrown", "188,143,143");
        add("RoyalBlue", "65,105,225");
        add("SaddleBrown", "139,69,19");
        add("Salmon", "250,128,114");
        add("SandyBrown", "244,164,96");
        add("SeaGreen", "46,139,87");
        add("SeaShell", "255,245,238");
        add("Sienna", "160,82,45");
        add("Silver", "192,192,192");
        add("SkyBlue", "135,206,235");
        add("SlateBlue", "106,90,205");
        add("SlateGray", "112,128,144");
        add("SlateGrey", "112,128,144");
        add("Snow", "255,250,250");
        add("SpringGreen", "0,255,127");
        add("SteelBlue", "70,130,180");
        add("Tan", "210,180,140");
        add("Teal", "0,128,128");
        add("Thistle", "216,191,216");
        add("Tomato", "255,99,71");
        add("Turquoise", "64,224,208");
        add("Violet", "238,130,238");
        add("Wheat", "245,222,179");
        add("White", "255,255,255");
        add("WhiteSmoke", "245,245,245");
        add("Yellow", "255,255,0");
        add("YellowGreen", "154,205,50");

        LIGHT_DEFAULTS = new Colors();
        DARK_DEFAULTS = new Colors();
        for (ThemeColor color : ALL) {
            LIGHT_DEFAULTS.setColor(color.getIndex(), color.getDefaultLight());
            DARK_DEFAULTS.setColor(color.getIndex(), color.getDefaultDark());
        }
        LIGHT_DEFAULTS.mReadOnly = true;
        DARK_DEFAULTS.mReadOnly = true;
        CURRENT = new Colors(LIGHT_DEFAULTS);
    }

    private Colors() {
        mColors = new Color[ALL.size()];
        for (int i = mColors.length - 1; i >= 0; i--) {
            mColors[i] = Color.BLACK;
        }
    }

    /**
     * Creates new theme colors from an existing ones.
     *
     * @param other The other theme colors to base this one off of.
     */
    public Colors(Colors other) {
        mColors = new Color[ALL.size()];
        System.arraycopy(other.mColors, 0, mColors, 0, mColors.length);
    }

    /**
     * Creates theme colors from a file.
     *
     * @param path The path to load the theme colors from.
     */
    public Colors(Path path) throws IOException {
        this(LIGHT_DEFAULTS);
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m       = Json.asMap(Json.parse(in));
            int     version = m.getInt(Settings.VERSION);
            if (version >= MINIMUM_VERSION && version <= CURRENT_VERSION && m.has(Settings.COLORS)) {
                load(m.getMap(Settings.COLORS));
            }
        }
    }

    /**
     * Creates theme colors from a JsonMap.
     *
     * @param m The map to load the theme colors from.
     */
    public Colors(JsonMap m) {
        this(LIGHT_DEFAULTS);
        load(m);
    }

    private void load(JsonMap m) {
        for (ThemeColor one : ALL) {
            if (m.has(one.getKey())) {
                setColor(one.getIndex(), decode(m.getString(one.getKey())));
            }
        }
    }

    /**
     * Save the theme colors to a file.
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
                w.key(Settings.COLORS);
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
     * Save the theme colors to a JsonWriter.
     *
     * @param w The JsonWriter to write to.
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        for (ThemeColor one : ALL) {
            w.keyValue(one.getKey(), encode(mColors[one.getIndex()]));
        }
        w.endMap();
    }

    /** @return {@code true} if this a stock theme and is therefore unmodifiable. */
    public boolean isReadOnly() {
        return mReadOnly;
    }

    /**
     * @param index The {@link ThemeColor} index to retrieve the {@link Color} for.
     * @return The {@link Color} for the index.
     */
    public Color getColor(int index) {
        if (index < 0 || index >= mColors.length) {
            return Color.BLACK;
        }
        return mColors[index];
    }

    /**
     * @param index The {@link ThemeColor} index to set the {@link Color} for.
     * @param color The new {@link Color} to use.
     */
    public void setColor(int index, Color color) {
        if (!mReadOnly && index >= 0 && index < mColors.length) {
            mColors[index] = new Color(color.getRGB(), true);
        }
    }

    /**
     * @param other The other colors to compare against.
     * @return {@code true} if these colors match the other colors.
     */
    public boolean match(Colors other) {
        int count = mColors.length;
        for (int i = 0; i < count; i++) {
            if (mColors[i].getRGB() != other.mColors[i].getRGB()) {
                return false;
            }
        }
        return true;
    }

    /** @return The default light theme colors. */
    public static Colors defaultLightThemeColors() {
        return LIGHT_DEFAULTS;
    }

    /** @return The default dark theme colors. */
    public static Colors defaultDarkThemeColors() {
        return DARK_DEFAULTS;
    }

    /** @return The current theme colors. */
    public static Colors currentThemeColors() {
        return CURRENT;
    }

    /**
     * Set the current theme colors.
     *
     * @param colors The theme colors to set as current.
     */
    public static void setCurrentThemeColors(Colors colors) {
        CURRENT = new Colors(colors);
    }

    private static void add(String name, String rgb) {
        NAME_TO_RGB.put(name.toLowerCase(), rgb);
        RGB_TO_NAME.put(rgb, name);
    }

    public static String encode(Color color) {
        StringBuilder buffer = new StringBuilder();
        buffer.append(color.getRed());
        buffer.append(',');
        buffer.append(color.getGreen());
        buffer.append(',');
        buffer.append(color.getBlue());
        int alpha = color.getAlpha();
        if (alpha != 255) {
            buffer.append(',');
            buffer.append(alpha);
        }
        String result     = buffer.toString();
        String substitute = RGB_TO_NAME.get(result);
        return substitute != null ? substitute : result;
    }

    public static Color decode(String buffer) {
        int red   = 0;
        int green = 0;
        int blue  = 0;
        int alpha = 0;
        if (buffer != null) {
            buffer = buffer.trim().toLowerCase();
            String value = NAME_TO_RGB.get(buffer);
            if (value != null) {
                buffer = value;
            }
            String[] color = buffer.split(",");
            if (color.length == 3 || color.length == 4) {
                red = parseColorComponent(color[0]);
                green = parseColorComponent(color[1]);
                blue = parseColorComponent(color[2]);
                alpha = color.length == 4 ? parseColorComponent(color[3]) : 255;
            } else if (color.length == 1 && color[0].startsWith("#")) {
                buffer = color[0].substring(1);
                int length = buffer.length();
                if (length == 3 || length == 4) {
                    StringBuilder newBuffer = new StringBuilder();
                    for (int i = 0; i < length; i++) {
                        char ch = buffer.charAt(i);
                        newBuffer.append(ch);
                        newBuffer.append(ch);
                    }
                    buffer = newBuffer.toString();
                }
                int single;
                try {
                    single = (int) (Long.parseLong(buffer, 16) & 0xFFFFFFFFL);
                } catch (NumberFormatException nfe) {
                    single = 0;
                }
                if (buffer.length() == 6) {
                    single |= 0xFF000000;
                }
                alpha = single >> 24 & 0xFF;
                red = single >> 16 & 0xFF;
                green = single >> 8 & 0xFF;
                blue = single & 0xFF;
            }
        }
        return new Color(red, green, blue, alpha);
    }

    private static int parseColorComponent(String buffer) {
        try {
            int value = Integer.parseInt(buffer);
            if (value < 0) {
                value = 0;
            } else if (value > 255) {
                value = 255;
            }
            return value;
        } catch (NumberFormatException nfe) {
            return 0;
        }
    }

    /**
     * @param color The color to check.
     * @return {@code true} if each color channel is the same.
     */
    public static boolean isMonochrome(Color color) {
        int green = color.getGreen();
        return color.getRed() == green && green == color.getBlue();
    }

    /**
     * @param color The color to check.
     * @return {@code true} if the color's perceived brightness is less than 50%.
     */
    public static boolean isDim(Color color) {
        return perceivedBrightness(color) < 0.5;
    }

    /**
     * @param color The color to check.
     * @return {@code true} if the color's perceived brightness is greater than or equal to 50%.
     */
    public static boolean isBright(Color color) {
        return perceivedBrightness(color) >= 0.5;
    }

    /**
     * @param color The color to check.
     * @return The perceived brightness. Less than 0.5 is a dark color.
     */
    public static double perceivedBrightness(Color color) {
        double red = color.getRed() / 255.0;
        if (!isMonochrome(color)) {
            double green = color.getGreen() / 255.0;
            double blue  = color.getBlue() / 255.0;
            return Math.sqrt(red * red * 0.241 + green * green * 0.691 + blue * blue * 0.068);
        }
        return red;
    }

    /**
     * @param color1     The first color.
     * @param color2     The second color.
     * @param percentage How much of the second color to use.
     * @return A color that is a blended version of the two passed in.
     */
    public static Color blend(Color color1, Color color2, int percentage) {
        int remaining = 100 - percentage;
        return new Color((color1.getRed() * remaining + color2.getRed() * percentage) / 100, (color1.getGreen() * remaining + color2.getGreen() * percentage) / 100, (color1.getBlue() * remaining + color2.getBlue() * percentage) / 100);
    }

    /**
     * @param color  The color to base the new color on.
     * @param amount The amount to adjust the saturation by, in the range -1 to 1.
     * @return The adjusted color.
     */
    public static Color adjustSaturation(Color color, float amount) {
        float[] hsb = Color.RGBtoHSB(color.getRed(), color.getGreen(), color.getBlue(), null);
        return new Color(Color.HSBtoRGB(hsb[0], Math.max(Math.min(hsb[1] + amount, 1.0f), 0.0f), hsb[2]));
    }

    /**
     * @param color  The color to base the new color on.
     * @param amount The amount to adjust the brightness by, in the range -1 to 1.
     * @return The adjusted color.
     */
    public static Color adjustBrightness(Color color, float amount) {
        float[] hsb = Color.RGBtoHSB(color.getRed(), color.getGreen(), color.getBlue(), null);
        return new Color(Color.HSBtoRGB(hsb[0], hsb[1], Math.max(Math.min(hsb[2] + amount, 1.0f), 0.0f)));
    }

    /**
     * @param color  The color to base the new color on.
     * @param amount The amount to adjust the hue by, in the range -1 to 1.
     * @return The adjusted color.
     */
    public static Color adjustHue(Color color, float amount) {
        float[] hsb = Color.RGBtoHSB(color.getRed(), color.getGreen(), color.getBlue(), null);
        return new Color(Color.HSBtoRGB(Math.max(Math.min(hsb[0] + amount, 1.0f), 0.0f), hsb[1], hsb[2]));
    }

    /**
     * @param color The color to work with.
     * @param alpha The alpha to use.
     * @return A new {@link Color} with the specified alpha.
     */
    public static Color getWithAlpha(Color color, int alpha) {
        return new Color(color.getRGB() & 0x00FFFFFF | alpha << 24, true);
    }
}
