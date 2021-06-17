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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.utility.I18n;

import java.awt.Graphics;
import javax.swing.SwingConstants;

/** A label that displays the "and" or the "or" message, or nothing if it is the first one. */
public class AndOrLabel extends Label {
    private Prereq mOwner;

    /**
     * Creates a new AndOrLabel.
     *
     * @param owner The owning {@link Prereq}.
     */
    public AndOrLabel(Prereq owner) {
        super(I18n.text("and"), SwingConstants.RIGHT);
        mOwner = owner;
    }

    @Override
    protected void paintComponent(Graphics gc) {
        PrereqList parent = mOwner.getParent();
        if (parent != null && parent.getChildren().get(0) != mOwner) {
            paintComponentWithText(gc, parent.requiresAll() ? I18n.text("and") : I18n.text("or"));
        }
    }
}
