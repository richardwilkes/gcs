/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.settings;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.FontIcon;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.Button;
import com.trollworks.gcs.ui.widget.KeyStrokeDisplay;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.SafeFileUpdater;
import com.trollworks.gcs.utility.VersionException;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Dimension;
import java.awt.event.InputEvent;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.swing.KeyStroke;

/** A window for editing menu key settings. */
public final class MenuKeySettingsWindow extends SettingsWindow<Map<String, String>> {
    private static MenuKeySettingsWindow INSTANCE;

    private static final String  NONE            = "NONE";
    private static final int     MINIMUM_VERSION = 1;
    private static final int     CURRENT_VERSION = 1;
    private static       boolean LOADED;

    private BandedPanel          mPanel;
    private Button               mResetButton;
    private Map<Button, Command> mMap;

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

    private MenuKeySettingsWindow() {
        super(I18n.text("Menu Key Settings"));
        mMap = new HashMap<>();
        fill();
    }

    @Override
    protected void preDispose() {
        synchronized (MenuKeySettingsWindow.class) {
            INSTANCE = null;
        }
    }

    @Override
    protected Panel createContent() {
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
        adjustForDuplicates();
        return mPanel;
    }

    private void addOne(Command cmd) {
        Button button = new Button(KeyStrokeDisplay.getKeyStrokeDisplay(KeyStroke.getKeyStroke('Z',
                InputEvent.META_DOWN_MASK | InputEvent.ALT_DOWN_MASK | InputEvent.CTRL_DOWN_MASK |
                        InputEvent.SHIFT_DOWN_MASK)), (btn) -> {
            Command          command = mMap.get(btn);
            KeyStrokeDisplay ksd     = new KeyStrokeDisplay(command);
            Modal dialog = Modal.prepareToShowMessage(this,
                    I18n.text("Type a keystroke…"), MessageType.QUESTION, ksd);
            Button resetButton = dialog.addButton(I18n.text("Reset"), 200);
            Button clearButton = dialog.addButton(I18n.text("Clear"), 100);
            dialog.addCancelButton();
            Button setButton = dialog.addButton(I18n.text("Set"), Modal.OK);
            ksd.setButtons(resetButton, clearButton, setButton);
            dialog.presentToUser();
            switch (dialog.getResult()) {
                case Modal.OK -> setAccelerator(btn, ksd.getKeyStroke());
                case 100 -> setAccelerator(btn, null); // Clear
                case 200 -> setAccelerator(btn, command.getOriginalAccelerator()); // Reset
            }
            adjustResetButton();
            adjustForDuplicates();
        });
        button.setThemeFont(Fonts.KEYBOARD);
        button.setBorder(new EmptyBorder(4));
        UIUtilities.setToPreferredSizeOnly(button);
        button.setText(KeyStrokeDisplay.getKeyStrokeDisplay(cmd.getAccelerator()));
        mMap.put(button, cmd);
        mPanel.add(button);
        mPanel.add(new Label(cmd.getTitle()));
    }

    private void setAccelerator(Button button, KeyStroke ks) {
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

    private void adjustForDuplicates() {
        Map<KeyStroke, Integer> counts = new HashMap<>();
        for (Map.Entry<Button, Command> entry : mMap.entrySet()) {
            KeyStroke ks = entry.getValue().getAccelerator();
            if (ks != null) {
                Integer count = counts.get(ks);
                if (count == null) {
                    counts.put(ks, Integer.valueOf(1));
                } else {
                    counts.put(ks, Integer.valueOf(1 + count.intValue()));
                }
            }
        }
        for (Map.Entry<Button, Command> entry : mMap.entrySet()) {
            int       count = 0;
            KeyStroke ks    = entry.getValue().getAccelerator();
            if (ks != null) {
                Integer c = counts.get(ks);
                if (c != null) {
                    count = c.intValue();
                }
            }
            Button button = entry.getKey();
            Label  label  = (Label) mPanel.getComponent(UIUtilities.getIndexOf(mPanel, button) + 1);
            if (count > 1) {
                label.setIcon(new FontIcon(FontAwesome.EXCLAMATION_TRIANGLE, Fonts.FONT_ICON_LABEL_PRIMARY, Colors.ERROR));
                label.setToolTipText(I18n.text("Duplicate key binding"));
            } else {
                label.setIcon(null);
                label.setToolTipText(null);
            }
        }
        mPanel.validate();
        mPanel.repaint();
    }

    @Override
    public void establishSizing() {
        pack();
        int width = getSize().width;
        setMinimumSize(new Dimension(width, 200));
        setMaximumSize(new Dimension(width, getPreferredSize().height));
    }

    @Override
    protected boolean shouldResetBeEnabled() {
        for (Command cmd : mMap.values()) {
            if (cmd.isOriginalAcceleratorOverridden()) {
                return true;
            }
        }
        return false;
    }

    @Override
    protected Map<String, String> getResetData() {
        Map<String, String> data = new HashMap<>();
        for (Command cmd : StdMenuBar.getCommands()) {
            KeyStroke ks = cmd.getOriginalAccelerator();
            data.put(cmd.getCommand(), ks != null ? ks.toString() : NONE);
        }
        return data;
    }

    @Override
    protected void doResetTo(Map<String, String> data) {
        for (Map.Entry<Button, Command> entry : mMap.entrySet()) {
            Command   cmd   = entry.getValue();
            String    value = data.get(cmd.getCommand());
            KeyStroke ks;
            if (value == null) {
                ks = cmd.getOriginalAccelerator();
            } else {
                if (NONE.equals(value)) {
                    ks = null;
                } else {
                    ks = KeyStroke.getKeyStroke(value);
                }
            }
            Button button = entry.getKey();
            setAccelerator(button, ks);
            button.invalidate();
        }
        adjustForDuplicates();
    }

    @Override
    protected Dirs getDir() {
        return Dirs.SETTINGS;
    }

    @Override
    protected FileType getFileType() {
        return FileType.KEY_SETTINGS;
    }

    @Override
    protected Map<String, String> createSettingsFrom(Path path) throws IOException {
        try (BufferedReader in = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            JsonMap m       = Json.asMap(Json.parse(in));
            int     version = m.getInt(DataFile.VERSION);
            if (version > CURRENT_VERSION) {
                throw VersionException.createTooNew();
            }
            if (version < MINIMUM_VERSION) {
                throw VersionException.createTooOld();
            }
            if (!m.has(Settings.KEY_BINDINGS)) {
                throw new IOException("invalid data type");
            }
            Map<String, String> data = new HashMap<>();
            m = m.getMap(Settings.KEY_BINDINGS);
            for (String key : m.keySet()) {
                data.put(key, m.getString(key));
            }
            return data;
        }
    }

    @Override
    protected void exportSettingsTo(Path path) throws IOException {
        SafeFileUpdater trans = new SafeFileUpdater();
        trans.begin();
        try {
            Files.createDirectories(path.getParent());
            File file = trans.getTransactionFile(path.toFile());
            try (JsonWriter w = new JsonWriter(new BufferedWriter(new FileWriter(file, StandardCharsets.UTF_8)), "\t")) {
                w.startMap();
                w.keyValue(Settings.VERSION, CURRENT_VERSION);
                w.key(Settings.KEY_BINDINGS);
                w.startMap();
                for (Command cmd : mMap.values()) {
                    KeyStroke ks = cmd.getAccelerator();
                    w.keyValue(cmd.getCommand(), ks == null ? NONE : ks.toString());
                }
                w.endMap();
                w.endMap();
            }
        } catch (IOException ioe) {
            trans.abort();
            throw ioe;
        }
        trans.commit();
    }
}
