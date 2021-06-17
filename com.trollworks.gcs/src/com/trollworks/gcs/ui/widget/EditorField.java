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

import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;

import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.text.ParseException;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;

/** Provides a standard editor field. */
public class EditorField extends JFormattedTextField implements ActionListener, Commitable {
    private ThemeFont mThemeFont;
    private String    mHint;

    public interface ChangeListener {
        void editorFieldChanged(EditorField field);
    }

    /**
     * Creates a new EditorField.
     *
     * @param formatter The formatter to use.
     * @param listener  The listener to use.
     * @param alignment The alignment to use.
     * @param value     The initial value.
     * @param tooltip   The tooltip to use.
     */
    public EditorField(AbstractFormatterFactory formatter, ChangeListener listener, int alignment, Object value, String tooltip) {
        this(formatter, listener, alignment, value, null, tooltip);
    }

    /**
     * Creates a new EditorField.
     *
     * @param formatter  The formatter to use.
     * @param listener   The listener to use.
     * @param alignment  The alignment to use.
     * @param value      The initial value.
     * @param protoValue The prototype value to use to set the preferred size.
     * @param tooltip    The tooltip to use.
     */
    public EditorField(AbstractFormatterFactory formatter, ChangeListener listener, int alignment, Object value, Object protoValue, String tooltip) {
        super(formatter, protoValue != null ? protoValue : value);
        setThemeFont(ThemeFont.FIELD_PRIMARY);
        setHorizontalAlignment(alignment);
        setToolTipText(tooltip);
        setFocusLostBehavior(COMMIT_OR_REVERT);
        setForeground(ThemeColor.ON_EDITABLE);
        setBackground(ThemeColor.EDITABLE);
        setCaretColor(ThemeColor.ON_EDITABLE);
        setSelectionColor(ThemeColor.SELECTION);
        setSelectedTextColor(ThemeColor.ON_SELECTION);
        setDisabledTextColor(ThemeColor.DISABLED_ON_EDITABLE);
        setBorder(new CompoundBorder(new LineBorder(ThemeColor.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
        if (protoValue != null) {
            setPreferredSize(getPreferredSize());
            setValue(value);
        }
        if (listener != null) {
            addPropertyChangeListener("value", (evt) -> listener.editorFieldChanged(this));
        }
        addActionListener(this);
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

    /** @param hint The hint to use, or {@code null}. */
    public void setHint(String hint) {
        mHint = hint;
        repaint();
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        super.paintComponent(gc);
        if (mHint != null && getText().isEmpty()) {
            Rectangle bounds = getBounds();
            bounds.x = 0;
            bounds.y = 0;
            gc.setColor(ThemeColor.HINT);
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
