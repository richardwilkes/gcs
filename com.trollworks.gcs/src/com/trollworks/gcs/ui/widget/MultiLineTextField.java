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
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeFont;

import java.awt.Dimension;
import java.awt.Font;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.event.FocusEvent;
import javax.swing.JTextArea;
import javax.swing.event.DocumentListener;

public class MultiLineTextField extends JTextArea {
    private ThemeFont mThemeFont;

    public MultiLineTextField(String text, String tooltip, DocumentListener listener) {
        super(text);
        setThemeFont(Fonts.FIELD_PRIMARY);
        setToolTipText(tooltip);
        setBorder(EditorField.createBorder(true));
        setFocusTraversalKeys(KeyboardFocusManager.FORWARD_TRAVERSAL_KEYS, KeyboardFocusManager.getCurrentKeyboardFocusManager().getDefaultFocusTraversalKeys(KeyboardFocusManager.FORWARD_TRAVERSAL_KEYS));
        setFocusTraversalKeys(KeyboardFocusManager.BACKWARD_TRAVERSAL_KEYS, KeyboardFocusManager.getCurrentKeyboardFocusManager().getDefaultFocusTraversalKeys(KeyboardFocusManager.BACKWARD_TRAVERSAL_KEYS));
        setFocusTraversalKeys(KeyboardFocusManager.UP_CYCLE_TRAVERSAL_KEYS, KeyboardFocusManager.getCurrentKeyboardFocusManager().getDefaultFocusTraversalKeys(KeyboardFocusManager.UP_CYCLE_TRAVERSAL_KEYS));
        setFocusTraversalKeysEnabled(true);
        setLineWrap(true);
        setWrapStyleWord(true);
        setForeground(Colors.ON_EDITABLE);
        setBackground(Colors.EDITABLE);
        setCaretColor(Colors.ON_EDITABLE);
        setSelectionColor(Colors.SELECTION);
        setSelectedTextColor(Colors.ON_SELECTION);
        setDisabledTextColor(new DynamicColor(() -> Colors.getWithAlpha(getForeground(), 96).getRGB()));
        setMinimumSize(new Dimension(50, 16));
        if (listener != null) {
            getDocument().addDocumentListener(listener);
        }
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        // For this to work as desired, we really want to have a known width. Unfortunately, the
        // AWT/Swing layout managers aren't setup that way, so we'll use our last width as a proxy
        // when possible. This may have undesired behaviors for other use-cases.
        if (getWidth() > 0) {
            return getPreferredSizeForWidth(getWidth());
        }
        Dimension size;
        String    text = getText();
        Font      font = getFont();
        if (text.isEmpty()) {
            size = new Dimension(0, TextDrawing.getFontHeight(font));
        } else {
            size = TextDrawing.getPreferredSize(font, text);
        }
        Insets insets = getInsets();
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    public Dimension getPreferredSizeForWidth(int width) {
        Dimension size;
        String    text   = getText();
        Font      font   = getFont();
        Insets    insets = getInsets();
        if (text.isEmpty()) {
            size = new Dimension(0, TextDrawing.getFontHeight(font));
        } else {
            size = TextDrawing.getPreferredSize(font,
                    TextDrawing.wrapToPixelWidth(font, getText(), width - (insets.left + insets.right)));
        }
        size.width += insets.left + insets.right;
        size.height += insets.top + insets.bottom;
        return size;
    }

    @Override
    protected void processFocusEvent(FocusEvent event) {
        super.processFocusEvent(event);
        if (isEnabled()) {
            if (event.getID() == FocusEvent.FOCUS_GAINED) {
                selectAll();
                setBorder(EditorField.createBorder(true));
                FocusHelper.scrollIntoView(this);
            } else {
                setBorder(EditorField.createBorder(false));
            }
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
