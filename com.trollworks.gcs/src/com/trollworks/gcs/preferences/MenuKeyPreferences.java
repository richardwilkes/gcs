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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.KeyStrokeDisplay;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.InputEvent;
import java.util.HashMap;
import java.util.Map;
import javax.swing.JButton;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JScrollPane;
import javax.swing.KeyStroke;

/** The menu keys preferences panel. */
public class MenuKeyPreferences extends PreferencePanel implements ActionListener {
    private static final String                    NONE = "NONE";
    private static       boolean                   LOADED;
    private              HashMap<JButton, Command> mMap = new HashMap<>();
    private              BandedPanel               mPanel;

    /**
     * Creates a new {@link MenuKeyPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public MenuKeyPreferences(PreferencesWindow owner) {
        super(I18n.Text("Menu Keys"), owner);
        setLayout(new BorderLayout());
        mPanel = new BandedPanel(I18n.Text("Menu Keys"));
        mPanel.setLayout(new ColumnLayout(2, 5, 0));
        mPanel.setBorder(new EmptyBorder(2, 5, 2, 5));
        mPanel.setOpaque(true);
        mPanel.setBackground(Color.WHITE);
        for (Command cmd : StdMenuBar.getCommands()) {
            JButton button = new JButton(KeyStrokeDisplay.getKeyStrokeDisplay(KeyStroke.getKeyStroke('Z', InputEvent.META_DOWN_MASK | InputEvent.ALT_DOWN_MASK | InputEvent.CTRL_DOWN_MASK | InputEvent.SHIFT_DOWN_MASK)));
            UIUtilities.setToPreferredSizeOnly(button);
            button.setText(getAcceleratorText(cmd));
            mMap.put(button, cmd);
            button.addActionListener(this);
            mPanel.add(button);
            JLabel label = new JLabel(cmd.getTitle());
            mPanel.add(label);
        }
        mPanel.setSize(mPanel.getPreferredSize());
        JScrollPane scroller      = new JScrollPane(mPanel);
        Dimension   preferredSize = scroller.getPreferredSize();
        if (preferredSize.height > 200) {
            preferredSize.height = 200;
        }
        scroller.setPreferredSize(preferredSize);
        add(scroller);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        JButton          button  = (JButton) event.getSource();
        Command          command = mMap.get(button);
        KeyStrokeDisplay ksd     = new KeyStrokeDisplay(command.getAccelerator());
        switch (WindowUtils.showOptionDialog(this, ksd, I18n.Text("Type a keystroke…"), false, JOptionPane.YES_NO_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, new Object[]{I18n.Text("Accept"), I18n.Text("Clear"), I18n.Text("Reset")}, null)) {
        case JOptionPane.CLOSED_OPTION:
        default:
            break;
        case JOptionPane.YES_OPTION: // Accept
            setAccelerator(button, ksd.getKeyStroke());
            break;
        case JOptionPane.NO_OPTION: // Clear
            setAccelerator(button, null);
            break;
        case JOptionPane.CANCEL_OPTION: // Reset
            setAccelerator(button, command.getOriginalAccelerator());
            break;
        }
        mPanel.setSize(mPanel.getPreferredSize());
        adjustResetButton();
    }

    @Override
    public void reset() {
        for (Map.Entry<JButton, Command> entry : mMap.entrySet()) {
            JButton button = entry.getKey();
            setAccelerator(button, entry.getValue().getOriginalAccelerator());
            button.invalidate();
        }
        mPanel.setSize(mPanel.getPreferredSize());
    }

    @Override
    public boolean isSetToDefaults() {
        for (Command cmd : mMap.values()) {
            if (!cmd.hasOriginalAccelerator()) {
                return false;
            }
        }
        return true;
    }

    private static String getAcceleratorText(Command cmd) {
        return KeyStrokeDisplay.getKeyStrokeDisplay(cmd.getAccelerator());
    }

    private void setAccelerator(JButton button, KeyStroke ks) {
        Command cmd = mMap.get(button);
        cmd.setAccelerator(ks);
        button.setText(getAcceleratorText(cmd));
        button.invalidate();
        Preferences prefs    = Preferences.getInstance();
        String      key      = cmd.getCommand();
        String      override = null;
        if (!cmd.hasOriginalAccelerator()) {
            override = ks != null ? ks.toString() : NONE;
        }
        prefs.setKeyBindingOverride(key, override);
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
