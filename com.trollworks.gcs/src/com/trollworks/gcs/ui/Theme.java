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

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Color;
import java.awt.Font;
import java.awt.Frame;
import java.io.IOException;

/**
 * Theme provides the colors used by GCS. Note it is expected to only be accessed on the UI thread.
 * If that expectation changes, the methods below will need to be synchronized.
 */
public class Theme {
    private static final String  KEY_FONTS      = "fonts";
    private static final String  KEY_NAME       = "name";
    private static final String  KEY_STYLE      = "style";
    private static final String  KEY_SIZE       = "size";
    private static final String  KEY_COLORS     = "colors";
    private static final Font    FALLBACK_FONT  = new Font(Font.DIALOG, Font.PLAIN, 12);
    private static final Color   FALLBACK_COLOR = Color.BLACK;
    /** The default theme. */
    public static final  Theme   DEFAULT;
    private static       Theme   CURRENT;
    private              Font[]  mFonts;
    private              Color[] mColors;
    private transient    boolean mReadOnly;

    static {
        DEFAULT = new Theme();
        for (ThemeFont font : ThemeFont.ALL) {
            DEFAULT.setFont(font.getIndex(), font.getDefault());
        }
        for (ThemeColor color : ThemeColor.ALL) {
            DEFAULT.setColor(color.getIndex(), color.getDefault());
        }
        DEFAULT.mReadOnly = true;
        CURRENT = new Theme(DEFAULT);
    }

    /** @return The current Theme. */
    public static Theme current() {
        return CURRENT;
    }

    /**
     * Set the current Theme.
     *
     * @param theme The Theme to set as the current Theme.
     */
    public static void set(Theme theme) {
        if (CURRENT != theme) {
            CURRENT = theme;
            repaint();
        }
    }

    /** Repaint all frames. */
    public static void repaint() {
        Frame[] frames = Frame.getFrames();
        for (Frame frame : frames) {
            if (frame.isShowing()) {
                frame.repaint();
            }
        }
    }

    private Theme() {
        mFonts = new Font[ThemeFont.ALL.size()];
        for (int i = mFonts.length - 1; i >= 0; i--) {
            mFonts[i] = FALLBACK_FONT;
        }
        mColors = new Color[ThemeColor.ALL.size()];
        for (int i = mColors.length - 1; i >= 0; i--) {
            mColors[i] = FALLBACK_COLOR;
        }
    }

    /**
     * Creates a new Theme from an existing Theme.
     *
     * @param other The other Theme to base this one off of.
     */
    public Theme(Theme other) {
        mFonts = new Font[ThemeFont.ALL.size()];
        System.arraycopy(other.mFonts, 0, mFonts, 0, mFonts.length);
        mColors = new Color[ThemeColor.ALL.size()];
        System.arraycopy(other.mColors, 0, mColors, 0, mColors.length);
    }

    /**
     * Creates a new Theme from a JsonMap.
     *
     * @param m The map to load the colors from.
     */
    public Theme(JsonMap m) {
        this(DEFAULT);
        if (m.has(KEY_FONTS)) {
            JsonMap m2 = m.getMap(KEY_FONTS);
            for (ThemeFont one : ThemeFont.ALL) {
                if (one.isEditable() && m2.has(one.getKey())) {
                    setFont(one.getIndex(), new FontDesc(m2.getMap(one.getKey())).create());
                }
            }
        }
        if (m.has(KEY_COLORS)) {
            JsonMap m2 = m.getMap(KEY_COLORS);
            for (ThemeColor one : ThemeColor.ALL) {
                if (m2.has(one.getKey())) {
                    setColor(one.getIndex(), Colors.decode(m2.getString(one.getKey())));
                }
            }
        }
    }

    /**
     * Save the Theme to a JsonWriter.
     *
     * @param w The JsonWriter to write to.
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();

        w.key(KEY_FONTS);
        w.startMap();
        for (ThemeFont one : ThemeFont.ALL) {
            if (one.isEditable()) {
                w.key(one.getKey());
                new FontDesc(getFont(one.getIndex())).toJSON(w);
            }
        }
        w.endMap();

        w.key(KEY_COLORS);
        w.startMap();
        for (ThemeColor one : ThemeColor.ALL) {
            w.keyValue(one.getKey(), Colors.encode(mColors[one.getIndex()]));
        }
        w.endMap();

        w.endMap();
    }

    /** @return {@code true} if this a stock Theme and is therefore unmodifiable. */
    public boolean isReadOnly() {
        return mReadOnly;
    }

    /**
     * @param index The {@link ThemeFont} index to retrieve the {@link Font} for.
     * @return The {@link Font} for the index.
     */
    public Font getFont(int index) {
        if (index < 0 || index >= mFonts.length) {
            return FALLBACK_FONT;
        }
        return mFonts[index];
    }

    /**
     * @param index The {@link ThemeFont} index to set the {@link Font} for.
     * @param font  The new {@link Font} to use.
     */
    public void setFont(int index, Font font) {
        if (!mReadOnly && index >= 0 && index < mFonts.length) {
            mFonts[index] = font;
        }
    }

    /**
     * @param index The {@link ThemeColor} index to retrieve the {@link Color} for.
     * @return The {@link Color} for the index.
     */
    public Color getColor(int index) {
        if (index < 0 || index >= mColors.length) {
            return FALLBACK_COLOR;
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
}
