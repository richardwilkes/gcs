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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.Graphics;
import javax.swing.JLabel;
import javax.swing.SwingConstants;

/** A label that displays the "or" message, or nothing if it is the first one. */
public class OrLabel extends JLabel {
    private Component mOwner;

    /**
     * Creates a new {@link OrLabel}.
     *
     * @param owner The owning component.
     */
    public OrLabel(Component owner) {
        super(I18n.Text("or"), SwingConstants.RIGHT);
        mOwner = owner;
        UIUtilities.setToPreferredSizeOnly(this);
    }

    @Override
    protected void paintComponent(Graphics gc) {
        setText(UIUtilities.getIndexOf(mOwner.getParent(), mOwner) != 0 ? I18n.Text("or") : "");
        super.paintComponent(GraphicsUtilities.prepare(gc));
    }
}
