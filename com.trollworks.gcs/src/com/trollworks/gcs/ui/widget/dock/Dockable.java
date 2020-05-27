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

package com.trollworks.gcs.ui.widget.dock;

import com.trollworks.gcs.ui.UIUtilities;

import java.awt.LayoutManager;
import javax.swing.Icon;
import javax.swing.JPanel;

/** Represents dockable items. */
public abstract class Dockable extends JPanel {
    /**
     * Creates a new {@link Dockable}.
     */
    protected Dockable() {
        super(true);
    }

    /** @param layout The {@link LayoutManager} to use. */
    protected Dockable(LayoutManager layout) {
        super(layout, true);
    }

    /** @return An {@link Icon} to represent this {@link Dockable}. */
    public abstract Icon getTitleIcon();

    /** @return The title of this {@link Dockable}. */
    public abstract String getTitle();

    /** @return The title tooltip of this {@link Dockable}. */
    public abstract String getTitleTooltip();

    /**
     * Called when this {@link Dockable} is made active within a {@link DockContainer}. This can be
     * called many times in a row without other {@link Dockable}s receiving a call in between.
     */
    public void activated() {
        // Does nothing by default
    }

    /** @return The containing {@link DockContainer}. */
    public final DockContainer getDockContainer() {
        return UIUtilities.getAncestorOfType(this, DockContainer.class);
    }
}
