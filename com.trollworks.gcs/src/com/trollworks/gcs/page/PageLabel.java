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

import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.widget.StdLabel;

import java.awt.Color;
import javax.swing.JComponent;

/** A label for a field in a page. */
public class PageLabel extends StdLabel {
    /**
     * Creates a new label.
     *
     * @param title The title of the field.
     */
    public PageLabel(String title) {
        this(title, ThemeColor.ON_CONTENT, null);
    }

    /**
     * Creates a new label for the specified field.
     *
     * @param title    The title of the field.
     * @param refersTo The component it refers to.
     */
    public PageLabel(String title, JComponent refersTo) {
        this(title, ThemeColor.ON_CONTENT, refersTo);
    }

    /**
     * Creates a new label for the specified field.
     *
     * @param title    The title of the field.
     * @param color    The color to use.
     * @param refersTo The component it refers to.
     */
    public PageLabel(String title, Color color, JComponent refersTo) {
        super(title);
        setThemeFont(ThemeFont.PAGE_LABEL_PRIMARY);
        setForeground(color);
        setRefersTo(refersTo);
    }
}
