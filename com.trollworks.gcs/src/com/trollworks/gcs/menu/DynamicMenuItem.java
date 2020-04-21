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
import javax.swing.JMenuItem;

/**
 * A replacement for {@link JMenuItem} that responds to changes in an attached {@link Action}'s
 * accelerator.
 */
public class DynamicMenuItem extends JMenuItem {
    /** Creates a new {@link DynamicMenuItem}. */
    public DynamicMenuItem() {
    }

    /**
     * Creates a new {@link DynamicMenuItem}.
     *
     * @param icon The icon to use.
     */
    public DynamicMenuItem(Icon icon) {
        super(icon);
    }

    /**
     * Creates a new {@link DynamicMenuItem}.
     *
     * @param text The text to use.
     */
    public DynamicMenuItem(String text) {
        super(text);
    }

    /**
     * Creates a new {@link DynamicMenuItem}.
     *
     * @param action The action to use.
     */
    public DynamicMenuItem(Action action) {
        super(action);
    }

    /**
     * Creates a new {@link DynamicMenuItem}.
     *
     * @param text The text to use.
     * @param icon The icon to use.
     */
    public DynamicMenuItem(String text, Icon icon) {
        super(text, icon);
    }

    /**
     * Creates a new {@link DynamicMenuItem}.
     *
     * @param text     The text to use.
     * @param mnemonic The mnemonic to use.
     */
    public DynamicMenuItem(String text, int mnemonic) {
        super(text, mnemonic);
    }

    @Override
    protected PropertyChangeListener createActionPropertyChangeListener(Action action) {
        return new DynamicJMenuItemPropertyChangeListener(this, action, super.createActionPropertyChangeListener(action));
    }
}
