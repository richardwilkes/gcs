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

package com.trollworks.gcs.ui.widget.dock;

import java.awt.Dimension;

/** A node within a {@link DockLayout}. */
public interface DockLayoutNode {
    /** @return The preferred size of this node. */
    Dimension getPreferredSize();

    /** @return The node's horizontal starting coordinate. */
    int getX();

    /** @return The node's vertical starting coordinate. */
    int getY();

    /** @return The node's width. */
    int getWidth();

    /** @return The node's height. */
    int getHeight();

    /**
     * Sets the position and size of this node, which may alter any contained sub-nodes.
     *
     * @param x      The horizontal starting coordinate.
     * @param y      The vertical starting coordinate.
     * @param width  The width;
     * @param height The height;
     */
    void setBounds(int x, int y, int width, int height);

    /** Invalidate then validate the layout of this node. */
    void revalidate();
}
