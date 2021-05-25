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

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.utility.PrintProxy;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.print.PageFormat;
import javax.swing.JPanel;

/** A printer page. */
public class Page extends JPanel {
    private PageOwner mOwner;

    /**
     * Creates a new page.
     *
     * @param owner The page owner.
     */
    public Page(PageOwner owner) {
        super(new BorderLayout());
        mOwner = owner;
        setOpaque(true);
        setBackground(ThemeColor.PAGE);
        PageFormat fmt    = mOwner.getPageSettings().createPageFormat();
        Insets     insets = mOwner.getPageAdornmentsInsets(this);
        setBorder(new EmptyBorder(insets.top + (int) fmt.getImageableY(), insets.left + (int) fmt.getImageableX(),
                insets.bottom + (int) (fmt.getHeight() - (fmt.getImageableY() + fmt.getImageableHeight())),
                insets.right + (int) (fmt.getWidth() - (fmt.getImageableX() + fmt.getImageableWidth()))));
        Scale     scale    = mOwner.getScale();
        Dimension pageSize = new Dimension(scale.scale((int) fmt.getWidth()), scale.scale((int) fmt.getHeight()));
        UIUtilities.setOnlySize(this, pageSize);
        setSize(pageSize);
    }

    @Override
    protected void paintComponent(Graphics gc) {
        super.paintComponent(GraphicsUtilities.prepare(gc));
        mOwner.drawPageAdornments(this, gc);
    }

    /**
     * @param component The {@link Component} to check.
     * @return If the {@link Component} or one of its ancestors is currently printing.
     */
    public static boolean isPrinting(Component component) {
        PrintProxy proxy = UIUtilities.getSelfOrAncestorOfType(component, PrintProxy.class);
        return proxy != null && proxy.isPrinting();
    }
}
