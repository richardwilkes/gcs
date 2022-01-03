/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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

import java.awt.Dimension;
import java.awt.Graphics;
import javax.swing.JComponent;
import javax.swing.plaf.ComponentUI;
import javax.swing.plaf.basic.BasicPopupMenuSeparatorUI;

public class ThemePopupMenuSeparatorUI extends BasicPopupMenuSeparatorUI {
    @SuppressWarnings("MethodOverridesStaticMethodOfSuperclass")
    public static ComponentUI createUI(JComponent c) {
        return new ThemePopupMenuSeparatorUI();
    }

    @Override
    public void paint(Graphics g, JComponent c) {
        Dimension s = c.getSize();
        g.setColor(Colors.DIVIDER);
        g.drawLine(0, 0, s.width, 0);
    }

    @Override
    public Dimension getPreferredSize(JComponent c) {
        return new Dimension(0, 1);
    }
}
