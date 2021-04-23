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
import java.awt.Frame;
import java.io.IOException;

/**
 * Theme provides the colors used by GCS. Note it is expected to only be accessed on the UI thread.
 * If that expectation changes, the methods below will need to be synchronized.
 */
public class Theme {
    /** The default theme. */
    public static final Theme   DEFAULT;
    private static      Theme   CURRENT;
    private             Color[] mColors;
    private             boolean mReadOnly;

    static {
        DEFAULT = new Theme();
        for (ThemeColor color : ThemeColor.ALL) {
            DEFAULT.setColor(color.getIndex(), color.getDefault());
        }
        DEFAULT.mReadOnly = true;
        CURRENT = new Theme(DEFAULT);
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
        mColors = new Color[ThemeColor.ALL.size()];
        for (int i = mColors.length - 1; i >= 0; i--) {
            mColors[i] = Color.BLACK;
        }
    }

    /**
     * Creates a new {@link Theme} from an existing {@link Theme}.
     *
     * @param other The other {@link Theme} to base this one off of.
     */
    public Theme(Theme other) {
        this();
        System.arraycopy(other.mColors, 0, mColors, 0, mColors.length);
    }

    /**
     * Creates a new {@link Theme} from a JsonMap.
     *
     * @param m The map to load the colors from.
     */
    public Theme(JsonMap m) {
        this(DEFAULT);
        for (ThemeColor one : ThemeColor.ALL) {
            String str = m.getString(one.getKey());
            if (!str.isBlank()) {
                setColor(one.getIndex(), Colors.decode(str));
            }
        }
    }

    /** @return {@code true} if this a stock {@link Theme} and is therefore unmodifiable. */
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
     * Save the {@link Theme} to a JsonWriter.
     *
     * @param w The JsonWriter to write to.
     */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        for (ThemeColor one : ThemeColor.ALL) {
            w.keyValue(one.getKey(), Colors.encode(mColors[one.getIndex()]));
        }
        w.endMap();
    }
}
