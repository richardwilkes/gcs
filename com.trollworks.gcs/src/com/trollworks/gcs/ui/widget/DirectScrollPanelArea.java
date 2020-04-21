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

package com.trollworks.gcs.ui.widget;

import java.awt.Point;

/** The possible {@link DirectScrollPanel} areas for mouse tracking. */
public enum DirectScrollPanelArea {
    /** In the content area. */
    CONTENT {
        @Override
        public void convertPoint(DirectScrollPanel owner, Point where) {
            owner.toContentView(where);
        }
    },
    /** In the header area. */
    HEADER {
        @Override
        public void convertPoint(DirectScrollPanel owner, Point where) {
            owner.toHeaderView(where);
        }
    },
    /** Outside any relevant areas. */
    NONE;

    /**
     * Adjusts the coordinates of the {@link Point} for the {@link DirectScrollPanelArea}.
     *
     * @param owner The owning {@link DirectScrollPanel}.
     * @param where The {@link Point} to convert.
     */
    public void convertPoint(DirectScrollPanel owner, Point where) {
        // Does nothing by default.
    }
}
