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

import java.awt.LayoutManager;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import javax.swing.JPanel;

/** A {@link JPanel} with {@link ActionListener} support. */
public class ActionPanel extends JPanel {
    private ArrayList<ActionListener> mActionListeners;
    private String                    mActionCommand;

    /** Creates a new {@link ActionPanel} with no layout. */
    public ActionPanel() {
        this(null);
    }

    /**
     * Creates a new {@link ActionPanel} with the specified layout.
     *
     * @param layout The layout manager to use. May be {@code null}.
     */
    public ActionPanel(LayoutManager layout) {
        super(layout);
    }

    /**
     * Adds an action listener.
     *
     * @param listener The listener to add.
     */
    public void addActionListener(ActionListener listener) {
        if (mActionListeners == null) {
            mActionListeners = new ArrayList<>(1);
        }
        if (!mActionListeners.contains(listener)) {
            mActionListeners.add(listener);
        }
    }

    /**
     * Removes an action listener.
     *
     * @param listener The listener to remove.
     */
    public void removeActionListener(ActionListener listener) {
        if (mActionListeners != null) {
            mActionListeners.remove(listener);
            if (mActionListeners.isEmpty()) {
                mActionListeners = null;
            }
        }
    }

    /** Notifies all action listeners with the standard action command. */
    public void notifyActionListeners() {
        notifyActionListeners(new ActionEvent(this, ActionEvent.ACTION_PERFORMED, getActionCommand()));
    }

    /**
     * Notifies all action listeners.
     *
     * @param event The action event to notify with.
     */
    public void notifyActionListeners(ActionEvent event) {
        if (mActionListeners != null && event != null) {
            for (ActionListener listener : new ArrayList<>(mActionListeners)) {
                listener.actionPerformed(event);
            }
        }
    }

    /** @return The action command. */
    public String getActionCommand() {
        return mActionCommand;
    }

    /** @param command The action command. */
    public void setActionCommand(String command) {
        mActionCommand = command;
    }
}
