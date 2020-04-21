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

import java.awt.Component;
import java.awt.KeyEventDispatcher;
import java.awt.KeyboardFocusManager;
import java.awt.event.KeyEvent;
import java.util.HashMap;
import javax.swing.Action;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.KeyStroke;
import javax.swing.event.MenuEvent;
import javax.swing.event.MenuListener;

/** Dynamically enables menu items just prior to them being used. */
public class DynamicMenuEnabler implements KeyEventDispatcher, MenuListener {
    private static final HashMap<KeyStroke, Command> MAP      = new HashMap<>();
    private static final DynamicMenuEnabler          INSTANCE = new DynamicMenuEnabler();

    static {
        KeyboardFocusManager.getCurrentKeyboardFocusManager().addKeyEventDispatcher(INSTANCE);
    }

    private DynamicMenuEnabler() {
        // Just here to prevent external instantiation
    }

    /** @param cmd The {@link Command} to add. */
    public static void add(Command cmd) {
        MAP.put(cmd.getAccelerator(), cmd);
    }

    /** @param cmd The {@link Command} to remove. */
    public static void remove(Command cmd) {
        MAP.remove(cmd.getAccelerator());
    }

    /** @param menu The {@link JMenu} to add. */
    public static void add(JMenu menu) {
        menu.addMenuListener(INSTANCE);
    }

    /** @param menu The {@link JMenu} to remove. */
    public static void remove(JMenu menu) {
        menu.removeMenuListener(INSTANCE);
    }

    @Override
    public boolean dispatchKeyEvent(KeyEvent event) {
        Command action = MAP.get(KeyStroke.getKeyStrokeForEvent(event));
        if (action != null) {
            action.adjust();
        }
        return false;
    }

    @Override
    public void menuCanceled(MenuEvent event) {
        // Not used.
    }

    @Override
    public void menuDeselected(MenuEvent event) {
        // Not used.
    }

    @Override
    public void menuSelected(MenuEvent event) {
        JMenu menu = (JMenu) event.getSource();
        for (Component component : menu.getMenuComponents()) {
            if (component instanceof JMenuItem) {
                JMenuItem item   = (JMenuItem) component;
                Action    action = item.getAction();
                if (action instanceof Command) {
                    ((Command) action).adjust();
                }
            }
        }
    }
}
