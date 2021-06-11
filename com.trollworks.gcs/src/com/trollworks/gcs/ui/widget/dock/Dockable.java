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

package com.trollworks.gcs.ui.widget.dock;

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.StdPanel;

import java.awt.LayoutManager;

/** Represents dockable items. */
public abstract class Dockable extends StdPanel {
    /**
     * Creates a new Dockable.
     */
    protected Dockable() {
    }

    /** @param layout The {@link LayoutManager} to use. */
    protected Dockable(LayoutManager layout) {
        super(layout);
    }

    /** @return An {@link RetinaIcon} to represent this Dockable. */
    public abstract RetinaIcon getTitleIcon();

    /** @return The title of this Dockable. */
    public abstract String getTitle();

    /** @return The title tooltip of this Dockable. */
    public abstract String getTitleTooltip();

    /**
     * Called when this Dockable is made active within a {@link DockContainer}. This can be called
     * many times in a row without other Dockables receiving a call in between.
     */
    public void activated() {
        // Does nothing by default
    }

    /** @return The containing {@link DockContainer}. */
    public final DockContainer getDockContainer() {
        return UIUtilities.getAncestorOfType(this, DockContainer.class);
    }
}
