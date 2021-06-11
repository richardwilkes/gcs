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
import com.trollworks.gcs.utility.text.Text;

import javax.swing.SwingConstants;

/** A header within the page. */
public class PageHeader extends StdLabel {
    /**
     * Creates a new PageHeader.
     *
     * @param title   The title to use.
     * @param tooltip The tooltip to use.
     */
    public PageHeader(String title, String tooltip) {
        super(title);
        setThemeFont(ThemeFont.PAGE_LABEL_PRIMARY);
        setHorizontalAlignment(SwingConstants.CENTER);
        setForeground(ThemeColor.ON_HEADER);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
    }
}
