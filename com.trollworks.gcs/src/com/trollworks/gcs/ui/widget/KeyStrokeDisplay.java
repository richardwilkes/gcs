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

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.utility.Platform;
import static java.awt.event.KeyEvent.VK_ALT;
import static java.awt.event.KeyEvent.VK_CAPS_LOCK;
import static java.awt.event.KeyEvent.VK_CONTROL;
import static java.awt.event.KeyEvent.VK_ESCAPE;
import static java.awt.event.KeyEvent.VK_META;
import static java.awt.event.KeyEvent.VK_SHIFT;
import static java.awt.event.KeyEvent.getKeyText;

import java.awt.Color;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import javax.swing.JLabel;
import javax.swing.KeyStroke;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.border.CompoundBorder;

/** Displays and captures keystrokes typed. */
public class KeyStrokeDisplay extends JLabel implements KeyListener {
    private KeyStroke mKeyStroke;

    /**
     * Creates a new {@link KeyStrokeDisplay}.
     *
     * @param ks The {@link KeyStroke} to start with.
     */
    public KeyStrokeDisplay(KeyStroke ks) {
        super(getKeyStrokeDisplay(KeyStroke.getKeyStroke('Z', InputEvent.META_DOWN_MASK | InputEvent.ALT_DOWN_MASK | InputEvent.CTRL_DOWN_MASK | InputEvent.SHIFT_DOWN_MASK)), SwingConstants.CENTER);
        setOpaque(true);
        setBackground(Color.WHITE);
        setBorder(new CompoundBorder(new LineBorder(), new EmptyBorder(2, 5, 2, 5)));
        addKeyListener(this);
        mKeyStroke = ks;
        UIUtilities.setToPreferredSizeOnly(this);
        setText(getKeyStrokeDisplay(mKeyStroke));
    }

    @Override
    public void keyPressed(KeyEvent event) {
        KeyStroke ks   = KeyStroke.getKeyStrokeForEvent(event);
        int       code = ks.getKeyCode();
        if (code != VK_SHIFT && code != VK_CONTROL && code != VK_META && code != VK_ALT && code != VK_CAPS_LOCK && code != VK_ESCAPE) {
            mKeyStroke = ks;
            setText(getKeyStrokeDisplay(mKeyStroke));
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
                    buffer.append(modifierText.replaceAll("\\+", ""));
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
        return buffer.toString();
    }
}
