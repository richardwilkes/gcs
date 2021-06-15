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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.utility.Platform;
import static java.awt.event.KeyEvent.VK_ALT;
import static java.awt.event.KeyEvent.VK_ALT_GRAPH;
import static java.awt.event.KeyEvent.VK_CAPS_LOCK;
import static java.awt.event.KeyEvent.VK_CONTROL;
import static java.awt.event.KeyEvent.VK_ESCAPE;
import static java.awt.event.KeyEvent.VK_META;
import static java.awt.event.KeyEvent.VK_SHIFT;
import static java.awt.event.KeyEvent.getKeyText;

import java.awt.Font;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.util.regex.Pattern;
import javax.swing.KeyStroke;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.border.CompoundBorder;

/** Displays and captures keystrokes typed. */
public class KeyStrokeDisplay extends StdLabel implements KeyListener, FocusListener {
    public static final  Font      KEYBOARD_FONT = new Font("Dialog", Font.PLAIN, 13);
    private static final Pattern   PLUS_PATTERN  = Pattern.compile("\\+");
    private              KeyStroke mKeyStroke;

    /**
     * Creates a new KeyStrokeDisplay.
     *
     * @param ks The {@link KeyStroke} to start with.
     */
    public KeyStrokeDisplay(KeyStroke ks) {
        super(getKeyStrokeDisplay(KeyStroke.getKeyStroke('Z', InputEvent.META_DOWN_MASK | InputEvent.ALT_DOWN_MASK | InputEvent.CTRL_DOWN_MASK | InputEvent.SHIFT_DOWN_MASK)), SwingConstants.CENTER);
        setThemeFont(null);
        setFont(KEYBOARD_FONT);
        setOpaque(true);
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
        setFocusable(true);
        addFocusListener(this);
        addKeyListener(this);
        mKeyStroke = ks;
        setText(getKeyStrokeDisplay(mKeyStroke));
    }

    @Override
    protected void setStdColors() {
        setBackground(ThemeColor.EDITABLE);
        setForeground(ThemeColor.ON_EDITABLE);
    }

    @Override
    public void keyPressed(KeyEvent event) {
        KeyStroke ks   = KeyStroke.getKeyStrokeForEvent(event);
        int       code = ks.getKeyCode();
        if (code != VK_SHIFT &&
                code != VK_CONTROL &&
                code != VK_META &&
                code != VK_ALT &&
                code != VK_ALT_GRAPH &&
                code != VK_CAPS_LOCK &&
                code != VK_ESCAPE) {
            mKeyStroke = ks;
            setText(getKeyStrokeDisplay(mKeyStroke));
            repaint();
        }
    }

    @Override
    public void keyReleased(KeyEvent event) {
        // Not used.
    }

    @Override
    public void keyTyped(KeyEvent event) {
        // Not used.
    }

    /** @return The {@link KeyStroke}. */
    public KeyStroke getKeyStroke() {
        return mKeyStroke;
    }

    /**
     * @param ks The {@link KeyStroke} to use.
     * @return The text that represents the {@link KeyStroke}.
     */
    public static String getKeyStrokeDisplay(KeyStroke ks) {
        StringBuilder buffer = new StringBuilder();
        if (ks != null) {
            int modifiers = ks.getModifiers();
            if (modifiers > 0) {
                String modifierText = InputEvent.getModifiersExText(modifiers);
                if (Platform.isMacintosh()) {
                    buffer.append(PLUS_PATTERN.matcher(modifierText).replaceAll(""));
                } else {
                    buffer.append(modifierText);
                    String delimiter = UIManager.getString("MenuItem.acceleratorDelimiter");
                    if (delimiter == null) {
                        delimiter = "+";
                    }
                    buffer.append(delimiter);
                }
            }
            int keyCode = ks.getKeyCode();
            if (keyCode == 0) {
                buffer.append(ks.getKeyChar());
            } else {
                buffer.append(getKeyText(keyCode));
            }
        }
        if (buffer.isEmpty()) {
            buffer.append(" ");
        }
        return buffer.toString();
    }

    @Override
    public void focusGained(FocusEvent event) {
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.ACTIVE_EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
    }

    @Override
    public void focusLost(FocusEvent event) {
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
    }
}
