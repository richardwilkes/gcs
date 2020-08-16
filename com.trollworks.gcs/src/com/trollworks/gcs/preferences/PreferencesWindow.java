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

import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.FlowLayout;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.WindowEvent;
import javax.swing.JButton;
import javax.swing.JPanel;
import javax.swing.JTabbedPane;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

/** A window for editing application preferences. */
public class PreferencesWindow extends BaseWindow implements ActionListener, ChangeListener, CloseHandler {
    private static PreferencesWindow INSTANCE;
    private        JTabbedPane       mTabPanel;
    private        JButton           mResetButton;

    /** Displays the preferences window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            PreferencesWindow wnd;
            synchronized (PreferencesWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new PreferencesWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private PreferencesWindow() {
        super(I18n.Text("Preferences"));
        mTabPanel = new JTabbedPane();
        addTab(new SheetPreferences(this));
        addTab(new DisplayPreferences(this));
        addTab(new OutputPreferences(this));
        addTab(new ThemePreferences(this));
        addTab(new MenuKeyPreferences(this));
        addTab(new ReferenceLookupPreferences(this));
        mTabPanel.addChangeListener(this);
        Container content = getContentPane();
        content.add(mTabPanel);
        content.add(createResetPanel(), BorderLayout.SOUTH);
        adjustResetButton();
        restoreBounds();
    }

    private void addTab(PreferencePanel panel) {
        mTabPanel.addTab(panel.toString(), panel);
    }

    @Override
    public void dispose() {
        synchronized (PreferencesWindow.class) {
            INSTANCE = null;
        }
        super.dispose();
    }

    private JPanel createResetPanel() {
        JPanel panel = new JPanel(new FlowLayout(FlowLayout.CENTER));
        mResetButton = new JButton(I18n.Text("Reset to Factory Defaults"));
        mResetButton.addActionListener(this);
        panel.add(mResetButton);
        return panel;
    }

    /** Call to adjust the reset button to the current panel. */
    public void adjustResetButton() {
        mResetButton.setEnabled(!((PreferencePanel) mTabPanel.getSelectedComponent()).isSetToDefaults());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ((PreferencePanel) mTabPanel.getSelectedComponent()).reset();
        adjustResetButton();
    }

    @Override
    public String getWindowPrefsKey() {
        return "preferences";
    }

    @Override
    public void stateChanged(ChangeEvent event) {
        adjustResetButton();
    }

    @Override
    public boolean mayAttemptClose() {
        return true;
    }

    @Override
    public boolean attemptClose() {
        windowClosing(new WindowEvent(this, WindowEvent.WINDOW_CLOSING));
        return true;
    }
}
