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

import java.awt.Graphics;
import java.awt.LayoutManager;
import javax.swing.JPanel;

public class StdPanel extends JPanel {
    public StdPanel() {
        init(true);
    }

    public StdPanel(LayoutManager layout) {
        super(layout);
        init(true);
    }

    public StdPanel(LayoutManager layout, boolean opaque) {
        super(layout);
        init(opaque);
    }

    private void init(boolean opaque) {
        setStdColors();
        setOpaque(opaque);
        setBorder(null);
    }

    protected void setStdColors() {
        setBackground(ThemeColor.BACKGROUND);
        setForeground(ThemeColor.ON_BACKGROUND);
    }

    protected void paintComponent(Graphics g) {
        if (isOpaque()) {
            g.setColor(getBackground());
            g.fillRect(0, 0, getWidth(), getHeight());
        }
        g.setColor(getForeground());
    }
}
