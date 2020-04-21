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

package com.trollworks.gcs.ui.widget;

import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.LayoutManager;
import javax.accessibility.Accessible;
import javax.swing.JComponent;
import javax.swing.JOptionPane;
import javax.swing.plaf.OptionPaneUI;
import javax.swing.plaf.basic.BasicOptionPaneUI;

/** A wrapper around the {@link BasicOptionPaneUI} that respects component minimum sizes. */
public class SizeAwareBasicOptionPaneUI extends BasicOptionPaneUI {
    private OptionPaneUI mOriginal;

    /**
     * Creates a new {@link SizeAwareBasicOptionPaneUI}.
     *
     * @param original The original {@link OptionPaneUI}.
     */
    public SizeAwareBasicOptionPaneUI(OptionPaneUI original) {
        mOriginal = original;
    }

    @Override
    public Dimension getMinimumSize(JComponent c) {
        if (c == optionPane) {
            Dimension     ourMin = getMinimumOptionPaneSize();
            LayoutManager lm     = c.getLayout();
            if (lm != null) {
                Dimension lmSize = lm.minimumLayoutSize(c);
                if (ourMin != null) {
                    return new Dimension(Math.max(lmSize.width, ourMin.width), Math.max(lmSize.height, ourMin.height));
                }
                return lmSize;
            }
            return ourMin;
        }
        return null;
    }

    @Override
    public boolean containsCustomComponents(JOptionPane op) {
        return mOriginal.containsCustomComponents(op);
    }

    @Override
    public void selectInitialValue(JOptionPane op) {
        mOriginal.selectInitialValue(op);
    }

    @Override
    public boolean contains(JComponent c, int x, int y) {
        return mOriginal.contains(c, x, y);
    }

    @Override
    public Accessible getAccessibleChild(JComponent c, int i) {
        return mOriginal.getAccessibleChild(c, i);
    }

    @Override
    public int getAccessibleChildrenCount(JComponent c) {
        return mOriginal.getAccessibleChildrenCount(c);
    }

    @Override
    public Dimension getMaximumSize(JComponent c) {
        return mOriginal.getMaximumSize(c);
    }

    @Override
    public Dimension getPreferredSize(JComponent c) {
        return mOriginal.getPreferredSize(c);
    }

    @Override
    public void installUI(JComponent c) {
        mOriginal.installUI(c);
    }

    @Override
    public void paint(Graphics g, JComponent c) {
        mOriginal.paint(g, c);
    }

    @Override
    public void uninstallUI(JComponent c) {
        mOriginal.uninstallUI(c);
    }

    @Override
    public void update(Graphics g, JComponent c) {
        mOriginal.update(g, c);
    }
}
