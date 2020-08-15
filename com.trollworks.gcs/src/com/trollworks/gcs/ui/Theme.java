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

import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.awt.Color;
import java.awt.Frame;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * Theme provides the colors used by GCS. Note it is expected to only be accessed on the UI thread.
 * If that expectation changes, the methods below will need to be synchronized.
 */
public class Theme implements Comparable<Theme>, Cloneable {
    /** The stock "light" theme (also the default). */
    public static final  Theme       LIGHT;
    /** The stock "dark" theme. */
    public static final  Theme       DARK;
    private static       Theme       CURRENT;
    private static final List<Theme> AVAILABLE;
    private static final int         CURRENT_VERSION = 1;
    private static final String      KEY_NAME        = "name";
    private static final String      KEY_COLORS      = "colors";
    private              String      mName;
    private              Color[]     mColors;
    private              boolean     mReadOnly;

    static {
        AVAILABLE = new ArrayList<>();

        LIGHT = new Theme(I18n.Text("Light"));
        LIGHT.setColor(ThemeColor.BANDING.getIndex(), new Color(232, 255, 232));
        LIGHT.setColor(ThemeColor.CURRENT_ENCUMBRANCE.getIndex(), new Color(252, 242, 196));
        LIGHT.setColor(ThemeColor.DIVIDER.getIndex(), Color.LIGHT_GRAY);
        LIGHT.setColor(ThemeColor.EDITABLE_MARKER.getIndex(), Color.LIGHT_GRAY);
        LIGHT.setColor(ThemeColor.ON_PAGE.getIndex(), Color.BLACK);
        LIGHT.setColor(ThemeColor.ON_USER_EDITABLE.getIndex(), new Color(0, 0, 192));
        LIGHT.setColor(ThemeColor.PAGE.getIndex(), Color.WHITE);
        LIGHT.setColor(ThemeColor.PAGE_VOID.getIndex(), Color.LIGHT_GRAY);
        LIGHT.setColor(ThemeColor.WARN.getIndex(), new Color(255, 205, 210));
        LIGHT.mReadOnly = true;
        AVAILABLE.add(LIGHT);

        DARK = new Theme(I18n.Text("Dark"));
        DARK.setColor(ThemeColor.BANDING.getIndex(), new Color(50, 53, 55));
        DARK.setColor(ThemeColor.CURRENT_ENCUMBRANCE.getIndex(), new Color(89, 91, 24));
        DARK.setColor(ThemeColor.DIVIDER.getIndex(), Color.DARK_GRAY);
        DARK.setColor(ThemeColor.EDITABLE_MARKER.getIndex(), Color.DARK_GRAY);
        DARK.setColor(ThemeColor.ON_PAGE.getIndex(), Color.WHITE);
        DARK.setColor(ThemeColor.ON_USER_EDITABLE.getIndex(), new Color(0, 0, 192));
        DARK.setColor(ThemeColor.PAGE.getIndex(), new Color(43, 43, 43));
        DARK.setColor(ThemeColor.PAGE_VOID.getIndex(), Color.BLACK);
        DARK.setColor(ThemeColor.WARN.getIndex(), new Color(90, 16, 16));
        DARK.mReadOnly = true;
        AVAILABLE.add(DARK);

        CURRENT = LIGHT;

        // This next chunk is currently commented out, but will be necessary if we want the look and
        // feel to have its colors updated by the theme as well.
        //
        // Take a look at javax.swing.plaf.basic.BasicLookAndFeel to find relevant keys
        //
        //UIDefaults def = UIManager.getDefaults();
        //def.put("Button.background", ThemeColor.ButtonBackground);
        //def.put("Button.foreground", ThemeColor.ButtonForeground);
        //def.put("CheckBox.background", ThemeColor.ContentBackground);
        //def.put("CheckBox.foreground", ThemeColor.ContentForeground);
        //def.put("CheckBoxMenuItem.background", ThemeColor.ContentBackground);
        //def.put("CheckBoxMenuItem.foreground", ThemeColor.ContentForeground);
        //def.put("ComboBox.background", ThemeColor.ContentBackground);
        //def.put("ComboBox.foreground", ThemeColor.ContentForeground);
        //def.put("EditorPane.background", ThemeColor.ContentBackground);
        //def.put("EditorPane.foreground", ThemeColor.ContentForeground);
        //def.put("Label.background", ThemeColor.ContentBackground);
        //def.put("Label.foreground", ThemeColor.ContentForeground);
        //def.put("List.background", ThemeColor.ContentBackground);
        //def.put("List.foreground", ThemeColor.ContentForeground);
        //def.put("Menu.background", ThemeColor.ContentBackground);
        //def.put("Menu.foreground", ThemeColor.ContentForeground);
        //def.put("MenuBar.background", ThemeColor.ContentBackground);
        //def.put("MenuBar.foreground", ThemeColor.ContentForeground);
        //def.put("MenuItem.background", ThemeColor.ContentBackground);
        //def.put("MenuItem.foreground", ThemeColor.ContentForeground);
        //def.put("OptionPane.background", ThemeColor.ContentBackground);
        //def.put("OptionPane.foreground", ThemeColor.ContentForeground);
        //def.put("Panel.background", ThemeColor.ContentBackground);
        //def.put("Panel.foreground", ThemeColor.ContentForeground);
        //def.put("PopupMenu.background", ThemeColor.ContentBackground);
        //def.put("PopupMenu.foreground", ThemeColor.ContentForeground);
        //def.put("ScrollPane.background", ThemeColor.ContentBackground);
        //def.put("ScrollPane.foreground", ThemeColor.ContentForeground);
        //def.put("TabbedPane.background", ThemeColor.ContentBackground);
        //def.put("TabbedPane.contentAreaColor", ThemeColor.ContentBackground);
        //def.put("TabbedPane.foreground", ThemeColor.ContentForeground);
        //def.put("TextArea.background", ThemeColor.ContentBackground);
        //def.put("TextArea.foreground", ThemeColor.ContentForeground);
        //def.put("TextField.background", ThemeColor.ContentBackground);
        //def.put("TextField.foreground", ThemeColor.ContentForeground);
        //def.put("TextPane.background", ThemeColor.ContentBackground);
        //def.put("TextPane.foreground", ThemeColor.ContentForeground);
        //def.put("Viewport.background", ThemeColor.ContentBackground);
        //def.put("Viewport.foreground", ThemeColor.ContentForeground);
    }

    /** @return The available {@link Theme}s. */
    public static List<Theme> getAvailable() {
        return new ArrayList<>(AVAILABLE);
    }

    /** @param theme A {@link Theme} to add to the available {@link Theme}s. */
    public static void add(Theme theme) {
        if (!AVAILABLE.contains(theme)) {
            AVAILABLE.add(theme);
            Collections.sort(AVAILABLE);
        }
    }

    /** @param theme A {@link Theme} to remove from the available {@link Theme}s. */
    public static void remove(Theme theme) {
        if (theme != LIGHT && theme != DARK) {
            AVAILABLE.remove(theme);
        }
    }

    /** @return The current {@link Theme}. */
    public static Theme current() {
        return CURRENT;
    }

    /**
     * Set the current {@link Theme}.
     *
     * @param theme The {@link Theme} to set as the current {@link Theme}.
     */
    public static void set(Theme theme) {
        if (CURRENT != theme) {
            CURRENT = theme;
            Frame[] frames = Frame.getFrames();
            for (Frame frame : frames) {
                if (frame.isShowing()) {
                    frame.repaint();
                }
            }
        }
    }

    private Theme(String name) {
        mName = name;
        mColors = new Color[ThemeColor.ALL.size()];
        for (int i = mColors.length - 1; i >= 0; i--) {
            mColors[i] = Color.BLACK;
        }
    }

    /**
     * Creates a new {@link Theme} from a JsonMap.
     *
     * @param m The map to load the colors from.
     * @throws IOException
     */
    public Theme(JsonMap m) throws IOException {
        this(I18n.Text("Untitled Theme"));
        // not currently looking at the version
        mName = m.getStringWithDefault(KEY_NAME, mName);
        JsonMap cm = m.getMap(KEY_COLORS);
        if (cm != null) {
            for (ThemeColor one : ThemeColor.ALL) {
                String str = m.getString(one.getKey());
                if (str != null) {
                    setColor(one.getIndex(), Colors.decode(str));
                }
            }
        }
    }

    @SuppressWarnings("MethodDoesntCallSuperMethod")
    @Override
    public Theme clone() {
        Theme other = new Theme(mName);
        System.arraycopy(mColors, 0, other.mColors, 0, mColors.length);
        return other;
    }

    @Override
    public int compareTo(Theme other) {
        return NumericComparator.caselessCompareStrings(mName, other.mName);
    }

    /** @return {@code true} if this a stock {@link Theme} and is therefore unmodifiable. */
    public boolean isReadOnly() {
        return mReadOnly;
    }

    /** @return The name of this {@link Theme}. */
    public String getName() {
        return mName;
    }

    /** @param name The new name for the {@link Theme}. */
    public void setName(String name) {
        if (!mReadOnly) {
            mName = name;
        }
    }

    @Override
    public String toString() {
        return mName;
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
            mColors[index] = color;
        }
    }

    /**
     * Save the {@link Theme} to a JsonWriter.
     *
     * @param w The JsonWriter to write to.
     * @throws IOException
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(LoadState.ATTRIBUTE_VERSION, CURRENT_VERSION);
        w.keyValue(KEY_NAME, mName);
        w.key(KEY_COLORS);
        w.startMap();
        for (ThemeColor one : ThemeColor.ALL) {
            w.keyValue(one.getKey(), Colors.encode(mColors[one.getIndex()]));
        }
        w.endMap();
        w.endMap();
    }
}
