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

/** A factory for creating new {@link Dockable} instances. */
public interface DockableFactory {
    /**
     * @param descriptor A descriptor that can be used to create a {@link Dockable}.
     * @return The newly created {@link Dockable}.
     */
    Dockable createDockable(String descriptor);
}
