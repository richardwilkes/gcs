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

import javax.swing.SwingConstants;

/** A header within the page. */
public class PageHeader extends Label {
    /**
     * Creates a new PageHeader.
     *
     * @param title The title to use.
     */
    public PageHeader(String title) {
        super(title);
        setThemeFont(Fonts.PAGE_LABEL_PRIMARY);
        setHorizontalAlignment(SwingConstants.CENTER);
        setForeground(Colors.ON_HEADER);
    }

    /**
     * Creates a new PageHeader.
     *
     * @param title   The title to use.
     * @param tooltip The tooltip to use.
     */
    public PageHeader(String title, String tooltip) {
        this(title);
        setToolTipText(tooltip);
    }
}
