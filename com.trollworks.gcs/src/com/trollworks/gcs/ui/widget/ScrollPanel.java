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

import com.trollworks.gcs.ui.ThemeColor;

import java.awt.Component;
import javax.swing.JScrollPane;

public class ScrollPanel extends JScrollPane {
    public ScrollPanel(Component view) {
        super(view);
        setBorder(null);
        getViewport().setBackground(ThemeColor.BACKGROUND);
    }

    public ScrollPanel(Component header, Component view) {
        this(view);
        setColumnHeaderView(header);
    }
}
