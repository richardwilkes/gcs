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

package com.trollworks.gcs.ui.image;

import java.awt.Cursor;
import java.awt.Point;
import java.awt.Toolkit;

public class Cursors {
    public static final Cursor HORIZONTAL_RESIZE = create("horizontal_resize_cursor");
    public static final Cursor VERTICAL_RESIZE   = create("vertical_resize_cursor");

    private static Cursor create(String name) {
        Img img = Images.get(name);
        return Toolkit.getDefaultToolkit().createCustomCursor(img, new Point(img.getWidth() / 2, img.getHeight() / 2), name);
    }
}
