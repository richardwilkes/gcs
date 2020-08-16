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

package com.trollworks.gcs.page;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Label;

import java.awt.Color;
import javax.swing.JComponent;
import javax.swing.UIManager;

/** A label for a field in a page. */
public class PageLabel extends Label {
    /**
     * Creates a new label for the specified field.
     *
     * @param title The title of the field.
     * @param field The field.
     */
    public PageLabel(String title, JComponent field) {
        super(title);
        setFont(UIManager.getFont(Fonts.KEY_LABEL_PRIMARY));
        setRefersTo(field);
        UIUtilities.setToPreferredSizeOnly(this);
    }

    /**
     * Creates a new label for the specified field.
     *
     * @param title The title of the field.
     * @param color The color to use.
     * @param field The field.
     */
    public PageLabel(String title, Color color, JComponent field) {
        super(title);
        setFont(UIManager.getFont(Fonts.KEY_LABEL_PRIMARY));
        setForeground(color);
        setRefersTo(field);
        UIUtilities.setToPreferredSizeOnly(this);
    }
}
