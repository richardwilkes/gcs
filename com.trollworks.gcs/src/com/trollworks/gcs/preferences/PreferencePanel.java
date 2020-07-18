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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.ui.border.EmptyBorder;

import javax.swing.JPanel;

/** The abstract base class for all preference panels. */
public abstract class PreferencePanel extends JPanel {
    private String            mTitle;
    private PreferencesWindow mOwner;

    /**
     * Creates a new preference panel.
     *
     * @param title The title for this panel.
     * @param owner The owning {@link PreferencesWindow}.
     */
    public PreferencePanel(String title, PreferencesWindow owner) {
        setBorder(new EmptyBorder(5));
        setOpaque(false);
        mTitle = title;
        mOwner = owner;
    }

    /** @return The owner. */
    public PreferencesWindow getOwner() {
        return mOwner;
    }

    /** Resets this panel back to its defaults. */
    public abstract void reset();

    /** @return Whether the panel is currently set to defaults or not. */
    public abstract boolean isSetToDefaults();

    /** Call to adjust the reset button for any changes that have been made. */
    protected void adjustResetButton() {
        mOwner.adjustResetButton();
    }

    @Override
    public String toString() {
        return mTitle;
    }
}
