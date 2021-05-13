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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.utility.text.Text;

import javax.swing.JLabel;
import javax.swing.SwingConstants;

public final class WidgetHelpers {
    private WidgetHelpers() {
    }

    /**
     * Creates a right-aligned label.
     *
     * @param title   The title of the label.
     * @param tooltip The tooltip for the label.
     */
    public static JLabel createLabel(String title, String tooltip) {
        JLabel label = new JLabel(title, SwingConstants.RIGHT);
        label.setOpaque(false);
        label.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        return label;
    }
}
