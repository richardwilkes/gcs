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
import com.trollworks.gcs.utility.text.Text;

import java.awt.Dimension;
import java.awt.KeyboardFocusManager;
import java.awt.event.FocusEvent;
import javax.swing.JTextArea;
import javax.swing.UIManager;
import javax.swing.border.CompoundBorder;
import javax.swing.event.DocumentListener;

public class MultiLineTextField extends JTextArea {
    public MultiLineTextField(String text, String tooltip, DocumentListener listener) {
        super(text);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        setBorder(UIManager.getBorder("FormattedTextField.border"));
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
        setDisabledTextColor(ThemeColor.DISABLED_ON_EDITABLE);
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
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
}
