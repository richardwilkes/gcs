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

package com.trollworks.gcs.ui.layout;

import java.awt.Rectangle;

/** The possible alignments. */
public enum Alignment {
    /** The left/top alignment. */
    LEFT_TOP,
    /** The center alignment. */
    CENTER,
    /** The right/bottom alignment. */
    RIGHT_BOTTOM;

    /**
     * Positions the inner rectangle within the outer one using the specified alignments.
     *
     * @param outer               The outer rectangle.
     * @param inner               The inner rectangle.
     * @param horizontalAlignment The horizontal alignment.
     * @param verticalAlignment   The vertical alignment.
     * @return The inner rectangle, which has been adjusted.
     */
    public static Rectangle position(Rectangle outer, Rectangle inner, Alignment horizontalAlignment, Alignment verticalAlignment) {
        switch (horizontalAlignment) {
        case LEFT_TOP:
        default:
            inner.x = outer.x;
            break;
        case CENTER:
            inner.x = outer.x + (outer.width - inner.width) / 2;
            break;
        case RIGHT_BOTTOM:
            inner.x = outer.x + outer.width - inner.width;
            break;
        }
        switch (verticalAlignment) {
        case LEFT_TOP:
        default:
            inner.y = outer.y;
            break;
        case CENTER:
            inner.y = outer.y + (outer.height - inner.height) / 2;
            break;
        case RIGHT_BOTTOM:
            inner.y = outer.y + outer.height - inner.height;
            break;
        }
        return inner;
    }
}
