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

import com.trollworks.gcs.ui.UIUtilities;

import java.awt.Component;
import java.awt.KeyboardFocusManager;
import java.awt.Toolkit;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.InputEvent;
import java.util.Objects;
import javax.swing.AbstractAction;
import javax.swing.Action;
import javax.swing.Icon;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.JPopupMenu;
import javax.swing.KeyStroke;

/** Represents a command in the user interface. */
public abstract class Command extends AbstractAction implements Comparable<Command> {
    /** The standard command modifier for this platform. */
    public static final int       COMMAND_MODIFIER         = Toolkit.getDefaultToolkit().getMenuShortcutKeyMaskEx();
    /** The standard command modifier for this platform, plus the shift key. */
    public static final int       SHIFTED_COMMAND_MODIFIER = COMMAND_MODIFIER | InputEvent.SHIFT_DOWN_MASK;
    private             KeyStroke mOriginalAccelerator;

    /**
     * Creates a new {@link Command}.
     *
     * @param title   The title to use.
     * @param command The command to use.
     */
    public Command(String title, String command) {
        super(title);
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title   The title to use.
     * @param command The command to use.
     * @param icon    The icon to use.
     */
    public Command(String title, String command, Icon icon) {
        super(title, icon);
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title   The title to use.
     * @param command The command to use.
     * @param keyCode The key code to use. The platform's standard menu shortcut key will be
     *                specified as a modifier.
     */
    public Command(String title, String command, int keyCode) {
        super(title);
        setAccelerator(keyCode);
        mOriginalAccelerator = getAccelerator();
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title     The title to use.
     * @param command   The command to use.
     * @param keyCode   The key code to use.
     * @param modifiers The modifiers to use.
     */
    public Command(String title, String command, int keyCode, int modifiers) {
        super(title);
        setAccelerator(keyCode, modifiers);
        mOriginalAccelerator = getAccelerator();
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title   The title to use.
     * @param command The command to use.
     * @param icon    The icon to use.
     * @param keyCode The key code to use. The platform's standard menu shortcut key will be
     *                specified as a modifier.
     */
    public Command(String title, String command, Icon icon, int keyCode) {
        super(title, icon);
        setAccelerator(keyCode);
        mOriginalAccelerator = getAccelerator();
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title     The title to use.
     * @param command   The command to use.
     * @param icon      The icon to use.
     * @param keyCode   The key code to use.
     * @param modifiers The modifiers to use.
     */
    public Command(String title, String command, Icon icon, int keyCode, int modifiers) {
        super(title, icon);
        setAccelerator(keyCode, modifiers);
        mOriginalAccelerator = getAccelerator();
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title     The title to use.
     * @param command   The command to use.
     * @param keystroke The {@link KeyStroke} to use.
     */
    public Command(String title, String command, KeyStroke keystroke) {
        super(title);
        setAccelerator(keystroke);
        mOriginalAccelerator = getAccelerator();
        setCommand(command);
    }

    /**
     * Creates a new {@link Command}.
     *
     * @param title     The title to use.
     * @param command   The command to use.
     * @param icon      The icon to use.
     * @param keystroke The {@link KeyStroke} to use.
     */
    public Command(String title, String command, Icon icon, KeyStroke keystroke) {
        super(title, icon);
        setAccelerator(keystroke);
        mOriginalAccelerator = getAccelerator();
        setCommand(command);
    }

    /** Called to adjust the action prior to a menu being displayed or an action being used. */
    public abstract void adjust();

    @Override
    public abstract void actionPerformed(ActionEvent event);

    /** @return The {@link Command}'s title. */
    public final String getTitle() {
        Object value = getValue(NAME);
        return value != null ? value.toString() : null;
    }

    /** @param title The {@link Command}'s title. */
    public final void setTitle(String title) {
        putValue(NAME, title);
    }

    /** @return The {@link Command}'s command string. */
    public final String getCommand() {
        Object value = getValue(ACTION_COMMAND_KEY);
        return value != null ? value.toString() : null;
    }

    /** @param cmd The {@link Command}'s command string. */
    public final void setCommand(String cmd) {
        putValue(ACTION_COMMAND_KEY, cmd);
    }

    /** @return The current keyboard accelerator for this {@link Command}, or {@code null}. */
    public final KeyStroke getAccelerator() {
        Object value = getValue(ACCELERATOR_KEY);
        return value instanceof KeyStroke ? (KeyStroke) value : null;
    }

    /** @return The original keyboard accelerator for this {@link Command}, or {@code null}. */
    public final KeyStroke getOriginalAccelerator() {
        return mOriginalAccelerator;
    }

    /** @return Whether the original accelerator is still set. */
    public final boolean hasOriginalAccelerator() {
        return Objects.equals(getAccelerator(), mOriginalAccelerator);
    }

    /**
     * Sets the keyboard accelerator for this {@link Command}. The platform's standard menu shortcut
     * key will be specified as a modifier.
     *
     * @param keyCode The key code to use.
     */
    public final void setAccelerator(int keyCode) {
        setAccelerator(keyCode, COMMAND_MODIFIER);
    }

    /**
     * Sets the keyboard accelerator for this {@link Command}.
     *
     * @param keyCode   The key code to use.
     * @param modifiers The modifiers to use.
     */
    public final void setAccelerator(int keyCode, int modifiers) {
        setAccelerator(KeyStroke.getKeyStroke(keyCode, modifiers));
    }

    /**
     * Sets the keyboard accelerator for this {@link Command}.
     *
     * @param keystroke The {@link KeyStroke} to use.
     */
    public final void setAccelerator(KeyStroke keystroke) {
        DynamicMenuEnabler.remove(this);
        putValue(ACCELERATOR_KEY, keystroke);
        DynamicMenuEnabler.add(this);
    }

    /** Removes any previously set keyboard accelerator. */
    public final void removeAccelerator() {
        DynamicMenuEnabler.remove(this);
        putValue(ACCELERATOR_KEY, null);
    }

    /** @return Whether any associated menu item should be marked. */
    public final boolean isMarked() {
        return Boolean.TRUE.equals(getValue(Action.SELECTED_KEY));
    }

    /** @param marked Whether any associated menu item should be marked. */
    public final void setMarked(boolean marked) {
        putValue(Action.SELECTED_KEY, Boolean.valueOf(marked));
    }

    @Override
    public int compareTo(Command other) {
        if (other == null) {
            return 1;
        }
        String title = getTitle();
        if (title == null) {
            title = "";
        }
        String otherTitle = other.getTitle();
        if (otherTitle == null) {
            otherTitle = "";
        }
        int result = title.compareTo(otherTitle);
        if (result == 0) {
            int hc  = hashCode();
            int ohc = other.hashCode();
            if (hc > ohc) {
                return 1;
            }
            if (hc < ohc) {
                return -1;
            }
        }
        return result;
    }

    /** @return The current permanent focus owner. */
    public static final Component getFocusOwner() {
        return KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
    }

    /** @return The current target. */
    public static final <T> T getTarget(Class<T> type) {
        Component focus  = getFocusOwner();
        T         target = UIUtilities.getSelfOrAncestorOfType(focus, type);
        if (target == null) {
            RetargetableFocus retargeter = UIUtilities.getSelfOrAncestorOfType(focus, RetargetableFocus.class);
            if (retargeter != null) {
                target = UIUtilities.getSelfOrAncestorOfType(retargeter.getRetargetedFocus(), type);
            }
        }
        return target;
    }

    /** @return The current active window. */
    public static final Window getActiveWindow() {
        return KeyboardFocusManager.getCurrentKeyboardFocusManager().getActiveWindow();
    }

    /** @return The current focused window. */
    public static final Window getFocusedWindow() {
        return KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusedWindow();
    }

    public static void adjustMenuTree(JPopupMenu menu) {
        for (Component component : menu.getComponents()) {
            if (component instanceof JMenu) {
                adjustMenuTree((JMenu) component);
            } else if (component instanceof JMenuItem) {
                JMenuItem item   = (JMenuItem) component;
                Action    action = item.getAction();
                if (action instanceof Command) {
                    ((Command) action).adjust();
                }
            }
        }
    }

    public static void adjustMenuTree(JMenu menu) {
        for (Component component : menu.getMenuComponents()) {
            if (component instanceof JMenu) {
                adjustMenuTree((JMenu) component);
            } else if (component instanceof JMenuItem) {
                JMenuItem item   = (JMenuItem) component;
                Action    action = item.getAction();
                if (action instanceof Command) {
                    ((Command) action).adjust();
                }
            }
        }
    }
}
