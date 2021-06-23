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

package com.trollworks.gcs.page;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.widget.Label;

import java.awt.Color;

/** A label for a field in a page. */
public class PageLabel extends Label {
    /**
     * Creates a new label.
     *
     * @param title The title of the field.
     */
    public PageLabel(String title) {   // TODO: Re-check colors here?
        this(title, Colors.ON_CONTENT);
    }

    /**
     * Creates a new label.
     *
     * @param title   The title of the field.
     * @param tooltip The tooltip to use.
     */
    public PageLabel(String title, String tooltip) {
        this(title, Colors.ON_CONTENT);
        setToolTipText(tooltip);
    }

    /**
     * Creates a new label for the specified field.
     *
     * @param title The title of the field.
     * @param color The color to use.
     */
    public PageLabel(String title, Color color) {
        super(title);
        setThemeFont(Fonts.PAGE_LABEL_PRIMARY);
        setForeground(color);
    }
}
