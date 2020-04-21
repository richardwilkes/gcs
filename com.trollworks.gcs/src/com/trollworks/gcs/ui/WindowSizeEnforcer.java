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

import com.trollworks.gcs.ui.widget.WindowUtils;

import java.awt.Dimension;
import java.awt.Window;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;

/** Ensures a window is never resized below its minimum size or above its maximum size settings. */
public class WindowSizeEnforcer implements ComponentListener {
    /**
     * Monitors a window for resizing.
     *
     * @param window The window to monitor.
     */
    public static void monitor(Window window) {
        window.addComponentListener(new WindowSizeEnforcer());
    }

    @Override
    public void componentHidden(ComponentEvent event) {
        // Not used.
    }

    @Override
    public void componentMoved(ComponentEvent event) {
        // Not used.
    }

    @Override
    public void componentResized(ComponentEvent event) {
        enforce((Window) event.getSource());
    }

    @Override
    public void componentShown(ComponentEvent event) {
        WindowUtils.forceOnScreen((Window) event.getComponent());
    }

    /** @param window The window to enforce min/max size on. */
    public static void enforce(Window window) {
        Dimension origSize  = window.getSize();
        Dimension otherSize = window.getMinimumSize();
        int       width     = origSize.width;
        int       height    = origSize.height;

        if (width < otherSize.width) {
            width = otherSize.width;
        }
        if (height < otherSize.height) {
            height = otherSize.height;
        }
        otherSize = window.getMaximumSize();
        if (width > otherSize.width) {
            width = otherSize.width;
        }
        if (height > otherSize.height) {
            height = otherSize.height;
        }
        if (width != origSize.width || height != origSize.height) {
            window.setSize(width, height);
        }
    }
}
