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

package com.trollworks.gcs.menu;

import java.beans.PropertyChangeListener;
import javax.swing.Action;
import javax.swing.Icon;
import javax.swing.JCheckBoxMenuItem;

/**
 * A replacement for {@link JCheckBoxMenuItem} that responds to changes in an attached {@link
 * Action}'s accelerator.
 */
public class DynamicCheckBoxMenuItem extends JCheckBoxMenuItem {
    /** Creates a new {@link DynamicCheckBoxMenuItem}. */
    public DynamicCheckBoxMenuItem() {
    }

    /**
     * Creates a new {@link DynamicCheckBoxMenuItem}.
     *
     * @param icon The icon to use.
     */
    public DynamicCheckBoxMenuItem(Icon icon) {
        super(icon);
    }

    /**
     * Creates a new {@link DynamicCheckBoxMenuItem}.
     *
     * @param text The text to use.
     */
    public DynamicCheckBoxMenuItem(String text) {
        super(text);
    }

    /**
     * Creates a new {@link DynamicCheckBoxMenuItem}.
     *
     * @param action The action to use.
     */
    public DynamicCheckBoxMenuItem(Action action) {
        super(action);
    }

    /**
     * Creates a new {@link DynamicCheckBoxMenuItem}.
     *
     * @param text The text to use.
     * @param icon The icon to use.
     */
    public DynamicCheckBoxMenuItem(String text, Icon icon) {
        super(text, icon);
    }

    /**
     * Creates a new {@link DynamicCheckBoxMenuItem}.
     *
     * @param text    The text to use.
     * @param checked The initial state to use.
     */
    public DynamicCheckBoxMenuItem(String text, boolean checked) {
        super(text, checked);
    }

    /**
     * Creates a new {@link DynamicCheckBoxMenuItem}.
     *
     * @param text    The text to use.
     * @param icon    The icon to use.
     * @param checked The initial state to use.
     */
    public DynamicCheckBoxMenuItem(String text, Icon icon, boolean checked) {
        super(text, icon, checked);
    }

    @Override
    protected PropertyChangeListener createActionPropertyChangeListener(Action action) {
        return new DynamicJMenuItemPropertyChangeListener(this, action, super.createActionPropertyChangeListener(action));
    }
}
