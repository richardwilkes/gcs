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

import com.trollworks.gcs.ui.Colors;

import java.awt.LayoutManager;

public class ContentPanel extends Panel {
    public ContentPanel() {
    }

    public ContentPanel(LayoutManager layout) {
        super(layout);
    }

    public ContentPanel(LayoutManager layout, boolean opaque) {
        super(layout, opaque);
    }

    @Override
    protected void setStdColors() {
        setBackground(Colors.CONTENT);
        setForeground(Colors.ON_CONTENT);
    }
}
