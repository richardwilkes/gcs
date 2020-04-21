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

import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeListener;
import java.text.ParseException;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** Provides a standard editor field. */
public class EditorField extends JFormattedTextField implements ActionListener, Commitable {
    private String mHint;

    /**
     * Creates a new {@link EditorField}.
     *
     * @param formatter The formatter to use.
     * @param listener  The listener to use.
     * @param alignment The alignment to use.
     * @param value     The initial value.
     * @param tooltip   The tooltip to use.
     */
    public EditorField(AbstractFormatterFactory formatter, PropertyChangeListener listener, int alignment, Object value, String tooltip) {
        this(formatter, listener, alignment, value, null, tooltip);
    }

    /**
     * Creates a new {@link EditorField}.
     *
     * @param formatter  The formatter to use.
     * @param listener   The listener to use.
     * @param alignment  The alignment to use.
     * @param value      The initial value.
     * @param protoValue The prototype value to use to set the preferred size.
     * @param tooltip    The tooltip to use.
     */
    public EditorField(AbstractFormatterFactory formatter, PropertyChangeListener listener, int alignment, Object value, Object protoValue, String tooltip) {
        super(formatter, protoValue != null ? protoValue : value);
        setHorizontalAlignment(alignment);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        if (protoValue != null) {
            setPreferredSize(getPreferredSize());
            setValue(value);
        }
        if (listener != null) {
            addPropertyChangeListener("value", listener);
        }
        addActionListener(this);
        setFocusLostBehavior(COMMIT_OR_REVERT);

        // Reset the selection colors back to what is standard for text fields.
        // This is necessary, since (at least on the Mac) JFormattedTextField
        // has the wrong values by default.
        setCaretColor(UIManager.getColor("TextField.caretForeground"));
        setSelectionColor(UIManager.getColor("TextField.selectionBackground"));
        setSelectedTextColor(UIManager.getColor("TextField.selectionForeground"));
        setDisabledTextColor(UIManager.getColor("TextField.inactiveForeground"));
    }

    @Override
    protected void processFocusEvent(FocusEvent event) {
        super.processFocusEvent(event);
        if (event.getID() == FocusEvent.FOCUS_GAINED) {
            selectAll();
        }
    }

    /** @param hint The hint to use, or {@code null}. */
    public void setHint(String hint) {
        mHint = hint;
        repaint();
    }

    @Override
    protected void paintComponent(Graphics gc) {
        super.paintComponent(gc);
        if (mHint != null && getText().isEmpty()) {
            Rectangle bounds = getBounds();
            bounds.x = 0;
            bounds.y = 0;
            gc.setColor(Color.GRAY);
            TextDrawing.draw(gc, bounds, mHint, SwingConstants.CENTER, SwingConstants.CENTER);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        attemptCommit();
    }

    @Override
    public void attemptCommit() {
        try {
            commitEdit();
        } catch (ParseException exception) {
            invalidEdit();
        }
    }
}
