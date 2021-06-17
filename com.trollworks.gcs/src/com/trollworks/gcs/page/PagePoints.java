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

import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.utility.I18n;

/** A points field in a page. */
public class PagePoints extends Label {
    public PagePoints(int points) {
        super("[" + points + "]");
        setThemeFont(ThemeFont.PAGE_LABEL_SECONDARY);
        setToolTipText(I18n.text("Points spent"));
        UIUtilities.setToPreferredSizeOnly(this);
    }
}
