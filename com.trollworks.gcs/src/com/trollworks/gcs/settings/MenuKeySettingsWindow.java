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
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.BaseWindow;
import com.trollworks.gcs.ui.widget.KeyStrokeDisplay;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.StdButton;
import com.trollworks.gcs.ui.widget.StdDialog;
import com.trollworks.gcs.ui.widget.StdLabel;
import com.trollworks.gcs.ui.widget.StdPanel;
import com.trollworks.gcs.ui.widget.StdScrollPanel;
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
import javax.swing.KeyStroke;

/** A window for editing menu key settings. */
public final class MenuKeySettingsWindow extends BaseWindow implements CloseHandler {
    private static final String                  NONE = "NONE";
    private static       boolean                 LOADED;
    private static       MenuKeySettingsWindow   INSTANCE;
    private              BandedPanel             mPanel;
    private              StdButton               mResetButton;
    private              Map<StdButton, Command> mMap;

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
        super(I18n.text("Menu Key Settings"));
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
        Container content = getContentPane();
        content.add(new StdScrollPanel(mPanel), BorderLayout.CENTER);
        content.add(createResetPanel(), BorderLayout.SOUTH);
        adjustResetButton();
        establishSizing();
        WindowUtils.packAndCenterWindowOn(this, null);
    }

    @Override
    public void establishSizing() {
        pack();
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, getPreferredSize().height));
    }

    private void addOne(Command cmd) {
        StdButton button = new StdButton(KeyStrokeDisplay.getKeyStrokeDisplay(KeyStroke.getKeyStroke('Z', InputEvent.META_DOWN_MASK | InputEvent.ALT_DOWN_MASK | InputEvent.CTRL_DOWN_MASK | InputEvent.SHIFT_DOWN_MASK)), (btn) -> {
            Command          command = mMap.get(btn);
            KeyStrokeDisplay ksd     = new KeyStrokeDisplay(command.getAccelerator());
            StdDialog        dialog  = StdDialog.prepareToShowMessage(this, I18n.text("Type a keystroke…"), MessageType.QUESTION, ksd);
            dialog.addButton(I18n.text("Accept"), StdDialog.OK);
            dialog.addButton(I18n.text("Clear"), 100);
            dialog.addButton(I18n.text("Reset"), 200);
            dialog.presentToUser();
            switch (dialog.getResult()) {
            case StdDialog.OK:
                setAccelerator(btn, ksd.getKeyStroke());
                break;
            case 100: // Clear
                setAccelerator(btn, null);
                break;
            case 200: // Reset
                setAccelerator(btn, command.getOriginalAccelerator());
                break;
            default: // Close
                break;
            }
            adjustResetButton();
        });
        button.setThemeFont(ThemeFont.KEYBOARD);
        UIUtilities.setToPreferredSizeOnly(button);
        button.setText(KeyStrokeDisplay.getKeyStrokeDisplay(cmd.getAccelerator()));
        mMap.put(button, cmd);
        StdPanel wrapper = new StdPanel(new PrecisionLayout().setMargins(4), false);
        wrapper.add(button);
        mPanel.add(wrapper);
        mPanel.add(new StdLabel(cmd.getTitle()));
    }

    private void setAccelerator(StdButton button, KeyStroke ks) {
        Command cmd = mMap.get(button);
        cmd.setAccelerator(ks);
        button.setText(KeyStrokeDisplay.getKeyStrokeDisplay(cmd.getAccelerator()));
        button.invalidate();
        Settings prefs    = Settings.getInstance();
        String   key      = cmd.getCommand();
        String   override = null;
        if (cmd.isOriginalAcceleratorOverridden()) {
            override = ks != null ? ks.toString() : NONE;
        }
        prefs.setKeyBindingOverride(key, override);
    }

    private StdPanel createResetPanel() {
        mResetButton = new StdButton(I18n.text("Reset to Factory Defaults"), (btn) -> {
            for (Map.Entry<StdButton, Command> entry : mMap.entrySet()) {
                StdButton button = entry.getKey();
                setAccelerator(button, entry.getValue().getOriginalAccelerator());
                button.invalidate();
            }
            adjustResetButton();
        });
        StdPanel panel = new StdPanel(new FlowLayout(FlowLayout.CENTER));
        panel.add(mResetButton);
        return panel;
    }

    private void adjustResetButton() {
        boolean enabled = false;
        for (Command cmd : mMap.values()) {
            if (cmd.isOriginalAcceleratorOverridden()) {
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
            Settings prefs = Settings.getInstance();
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
