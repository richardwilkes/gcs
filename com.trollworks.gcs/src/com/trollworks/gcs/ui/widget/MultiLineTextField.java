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

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.DynamicColor;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;

import java.awt.Dimension;
import java.awt.Font;
import java.awt.KeyboardFocusManager;
import java.awt.event.FocusEvent;
import javax.swing.JTextArea;
import javax.swing.border.CompoundBorder;
import javax.swing.event.DocumentListener;

public class MultiLineTextField extends JTextArea {
    private ThemeFont mThemeFont;

    public MultiLineTextField(String text, String tooltip, DocumentListener listener) {
        super(text);
        setThemeFont(ThemeFont.FIELD_PRIMARY);
        setToolTipText(tooltip);
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
        setFocusTraversalKeys(KeyboardFocusManager.FORWARD_TRAVERSAL_KEYS, KeyboardFocusManager.getCurrentKeyboardFocusManager().getDefaultFocusTraversalKeys(KeyboardFocusManager.FORWARD_TRAVERSAL_KEYS));
        setFocusTraversalKeys(KeyboardFocusManager.BACKWARD_TRAVERSAL_KEYS, KeyboardFocusManager.getCurrentKeyboardFocusManager().getDefaultFocusTraversalKeys(KeyboardFocusManager.BACKWARD_TRAVERSAL_KEYS));
        setFocusTraversalKeys(KeyboardFocusManager.UP_CYCLE_TRAVERSAL_KEYS, KeyboardFocusManager.getCurrentKeyboardFocusManager().getDefaultFocusTraversalKeys(KeyboardFocusManager.UP_CYCLE_TRAVERSAL_KEYS));
        setFocusTraversalKeysEnabled(true);
        setLineWrap(true);
        setWrapStyleWord(true);
        setForeground(ThemeColor.ON_EDITABLE);
        setBackground(ThemeColor.EDITABLE);
        setCaretColor(ThemeColor.ON_EDITABLE);
        setSelectionColor(ThemeColor.SELECTION);
        setSelectedTextColor(ThemeColor.ON_SELECTION);
        setDisabledTextColor(new DynamicColor(() -> Colors.getWithAlpha(getForeground(), 96).getRGB()));
        setMinimumSize(new Dimension(50, 16));
        if (listener != null) {
            getDocument().addDocumentListener(listener);
        }
    }

    @Override
    protected void processFocusEvent(FocusEvent event) {
        super.processFocusEvent(event);
        if (event.getID() == FocusEvent.FOCUS_GAINED) {
            selectAll();
            setBorder(new CompoundBorder(new LineBorder(ThemeColor.ACTIVE_EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
        } else {
            setBorder(new CompoundBorder(new LineBorder(ThemeColor.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
        }
    }

    @Override
    public ToolTip createToolTip() {
        return new ToolTip(this);
    }

    public final ThemeFont getThemeFont() {
        return mThemeFont;
    }

    public final void setThemeFont(ThemeFont font) {
        mThemeFont = font;
    }

    @Override
    public final Font getFont() {
        if (mThemeFont == null) {
            // If this happens, we are in the constructor and the look & feel is being inited, so
            // just return whatever was there by default.
            return super.getFont();
        }
        return mThemeFont.getFont();
    }
}
