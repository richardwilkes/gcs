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

package com.trollworks.gcs.settings;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.KeyStrokeDisplay;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.event.InputEvent;
import java.awt.event.WindowEvent;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.JButton;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.KeyStroke;

/** A window for editing menu key settings. */
public final class MenuKeySettingsWindow extends BaseWindow implements CloseHandler {
    private static final String                NONE = "NONE";
    private static       boolean               LOADED;
    private static       MenuKeySettingsWindow INSTANCE;
    private              BandedPanel           mPanel;
    private              JButton               mResetButton;
    private              Map<JButton, Command> mMap;

    /** Displays the menu key settings window. */
    public static void display() {
        if (!UIUtilities.inModalState()) {
            MenuKeySettingsWindow wnd;
            synchronized (MenuKeySettingsWindow.class) {
                if (INSTANCE == null) {
                    INSTANCE = new MenuKeySettingsWindow();
                }
                wnd = INSTANCE;
            }
            wnd.setVisible(true);
        }
    }

    private MenuKeySettingsWindow() {
        super(I18n.Text("Menu Key Settings"));
        mMap = new HashMap<>();
        mPanel = new BandedPanel(true);
        mPanel.setLayout(new PrecisionLayout().setColumns(4).setMargins(0, 10, 0, 26).setVerticalSpacing(0));
        List<Command> cmds = StdMenuBar.getCommands();
        int           all  = cmds.size();
        int           half = all / 2;
        if (half + half != all) {
            half++;
        }
        for (int i = 0; i < half; i++) {
            addOne(cmds.get(i));
            if (i + half < all) {
                addOne(cmds.get(i + half));
            }
        }
        Container   content  = getContentPane();
        JScrollPane scroller = new JScrollPane(mPanel);
        scroller.setBorder(null);
        content.add(scroller, BorderLayout.CENTER);
        content.add(createResetPanel(), BorderLayout.SOUTH);
        adjustResetButton();
        WindowUtils.packAndCenterWindowOn(this, null);
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, getPreferredSize().height));
    }

    private void addOne(Command cmd) {
        JButton button = new JButton(KeyStrokeDisplay.getKeyStrokeDisplay(KeyStroke.getKeyStroke('Z', InputEvent.META_DOWN_MASK | InputEvent.ALT_DOWN_MASK | InputEvent.CTRL_DOWN_MASK | InputEvent.SHIFT_DOWN_MASK)));
        UIUtilities.setToPreferredSizeOnly(button);
        button.setText(KeyStrokeDisplay.getKeyStrokeDisplay(cmd.getAccelerator()));
        mMap.put(button, cmd);
        button.addActionListener((evt) -> {
            JButton          btn     = (JButton) evt.getSource();
            Command          command = mMap.get(btn);
            KeyStrokeDisplay ksd     = new KeyStrokeDisplay(command.getAccelerator());
            switch (WindowUtils.showOptionDialog(this, ksd, I18n.Text("Type a keystroke…"), false, JOptionPane.YES_NO_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, new Object[]{I18n.Text("Accept"), I18n.Text("Clear"), I18n.Text("Reset")}, null)) {
            case JOptionPane.CLOSED_OPTION:
            default:
                break;
            case JOptionPane.YES_OPTION: // Accept
                setAccelerator(btn, ksd.getKeyStroke());
                break;
            case JOptionPane.NO_OPTION: // Clear
                setAccelerator(btn, null);
                break;
            case JOptionPane.CANCEL_OPTION: // Reset
                setAccelerator(btn, command.getOriginalAccelerator());
                break;
            }
            adjustResetButton();
        });
        JPanel wrapper = new JPanel(new PrecisionLayout().setMargins(4));
        wrapper.setOpaque(false);
        wrapper.add(button);
        mPanel.add(wrapper);
        mPanel.add(new JLabel(cmd.getTitle()));
    }

    private void setAccelerator(JButton button, KeyStroke ks) {
        Command cmd = mMap.get(button);
        cmd.setAccelerator(ks);
        button.setText(KeyStrokeDisplay.getKeyStrokeDisplay(cmd.getAccelerator()));
        button.invalidate();
        Preferences prefs    = Preferences.getInstance();
        String      key      = cmd.getCommand();
        String      override = null;
        if (!cmd.hasOriginalAccelerator()) {
            override = ks != null ? ks.toString() : NONE;
        }
        prefs.setKeyBindingOverride(key, override);
    }

    private JPanel createResetPanel() {
        mResetButton = new JButton(I18n.Text("Reset to Factory Defaults"));
        mResetButton.addActionListener((evt) -> {
            for (Map.Entry<JButton, Command> entry : mMap.entrySet()) {
                JButton button = entry.getKey();
                setAccelerator(button, entry.getValue().getOriginalAccelerator());
                button.invalidate();
            }
            adjustResetButton();
        });
        JPanel panel = new JPanel(new FlowLayout(FlowLayout.CENTER));
        panel.add(mResetButton);
        return panel;
    }

    private void adjustResetButton() {
        boolean enabled = false;
        for (Command cmd : mMap.values()) {
            if (!cmd.hasOriginalAccelerator()) {
                enabled = true;
                break;
            }
        }
        mResetButton.setEnabled(enabled);
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

    @Override
    public void dispose() {
        synchronized (MenuKeySettingsWindow.class) {
            INSTANCE = null;
        }
        super.dispose();
    }

    /** Loads the current menu key settings from the preferences file. */
    public static synchronized void loadFromPreferences() {
        if (!LOADED) {
            Preferences prefs = Preferences.getInstance();
            for (Command cmd : StdMenuBar.getCommands()) {
                String value = prefs.getKeyBindingOverride(cmd.getCommand());
                if (value != null) {
                    if (NONE.equals(value)) {
                        cmd.setAccelerator(null);
                    } else {
                        cmd.setAccelerator(KeyStroke.getKeyStroke(value));
                    }
                }
            }
            LOADED = true;
        }
    }
}
