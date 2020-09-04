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

package com.trollworks.gcs.utility;

import java.awt.Rectangle;

/** Provides geometry-related utilities. */
public final class Geometry {
    private Geometry() {
    }

    /**
     * Intersects two {@link Rectangle}s, producing a third. Unlike the {@link
     * Rectangle#intersection(Rectangle)} method, the resulting {@link Rectangle}'s width & height
     * will not be set to less than zero when there is no overlap.
     *
     * @param first  The first {@link Rectangle}.
     * @param second The second {@link Rectangle}.
     * @return The intersection of the two {@link Rectangle}s.
     */
    public static Rectangle intersection(Rectangle first, Rectangle second) {
        if (first.width < 1 || first.height < 1 || second.width < 1 || second.height < 1) {
            return new Rectangle();
        }
        int x = Math.max(first.x, second.x);
        int y = Math.max(first.y, second.y);
        int w = Math.min(first.x + first.width, second.x + second.width) - x;
        int h = Math.min(first.y + first.height, second.y + second.height) - y;
        if (w < 0 || h < 0) {
            return new Rectangle();
        }
        return new Rectangle(x, y, w, h);
    }

    /**
     * Unions two {@link Rectangle}s, producing a third. Unlike the {@link
     * Rectangle#union(Rectangle)} method, an empty {@link Rectangle} will not cause the {@link
     * Rectangle}'s boundary to extend to the 0,0 point.
     *
     * @param first  The first {@link Rectangle}.
     * @param second The second {@link Rectangle}.
     * @return The resulting {@link Rectangle}.
     */
    public static Rectangle union(Rectangle first, Rectangle second) {
        boolean firstEmpty  = first.width < 1 || first.height < 1;
        boolean secondEmpty = second.width < 1 || second.height < 1;
        if (firstEmpty && secondEmpty) {
            return new Rectangle();
        }
        if (firstEmpty) {
            return new Rectangle(second);
        }
        if (secondEmpty) {
            return new Rectangle(first);
        }
        return first.union(second);
    }

    /**
     * @param amount The number of pixels to inset the {@link Rectangle}.
     * @param bounds The {@link Rectangle} to inset.
     * @return The {@link Rectangle} that was passed in.
     */
    public static Rectangle inset(int amount, Rectangle bounds) {
        bounds.x += amount;
        bounds.y += amount;
        bounds.width -= amount * 2;
        bounds.height -= amount * 2;
        if (bounds.width < 0) {
            bounds.width = 0;
        }
        if (bounds.height < 0) {
            bounds.height = 0;
        }
        return bounds;
    }
}
