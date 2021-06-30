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
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.TextDrawing;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.text.ParseException;
import java.util.Objects;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.border.Border;
import javax.swing.border.CompoundBorder;

/** Provides a standard editor field. */
public class EditorField extends JFormattedTextField implements ActionListener, Commitable {
    private ThemeFont mThemeFont;
    private String    mHint;
    private String    mErrorMsg;
    private String    mOriginalTooltip;
    private Object    mPrototypeValue;

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
        super(formatter, value);
        mPrototypeValue = protoValue;
        setThemeFont(Fonts.FIELD_PRIMARY);
        setHorizontalAlignment(alignment);
        setToolTipText(tooltip);
        setFocusLostBehavior(COMMIT_OR_REVERT);
        setForeground(Colors.ON_EDITABLE);
        setBackground(Colors.EDITABLE);
        setCaretColor(Colors.ON_EDITABLE);
        setSelectionColor(Colors.SELECTION);
        setSelectedTextColor(Colors.ON_SELECTION);
        setDisabledTextColor(new DynamicColor(() -> Colors.getWithAlpha(getForeground(), 96).getRGB()));
        setBorder(new CompoundBorder(new LineBorder(Colors.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4)));
        if (listener != null) {
            addPropertyChangeListener("value", (evt) -> listener.editorFieldChanged(this));
        }
        addActionListener(this);
    }

    @Override
    public Dimension getPreferredSize() {
        if (isPreferredSizeSet()) {
            return super.getPreferredSize();
        }
        if (mPrototypeValue != null) {
            // RAW: This is a horrible way to do this... but until I write my own text field, it
            //      will have to do.
            JFormattedTextField tmp = new JFormattedTextField(getFormatterFactory(), mPrototypeValue);
            tmp.setFont(getFont());
            tmp.setBorder(createBorder(isFocusOwner()));
            tmp.setHorizontalAlignment(getHorizontalAlignment());
            Dimension size = tmp.getPreferredSize();
            size.width += Scale.get(this).scale(4); // Add some slop, since it is being truncated otherwise
            return size;
        }
        return super.getPreferredSize();
    }

    private static Border createBorder(boolean focused) {
        return new CompoundBorder(new LineBorder(focused ? Colors.ACTIVE_EDITABLE_BORDER :
                Colors.EDITABLE_BORDER), new EmptyBorder(2, 4, 2, 4));
    }

    @Override
    protected void processFocusEvent(FocusEvent event) {
        super.processFocusEvent(event);
        boolean focused = event.getID() == FocusEvent.FOCUS_GAINED;
        if (focused) {
            selectAll();
        }
        setBorder(createBorder(focused));
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
            gc.setColor(Colors.HINT);
            Font font = getFont();
            gc.setFont(font);
            TextDrawing.draw(gc, bounds, TextDrawing.truncateIfNecessary(font, mHint,
                    bounds.width, SwingConstants.CENTER), SwingConstants.CENTER,
                    SwingConstants.CENTER);
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

    @Override
    public void setToolTipText(String text) {
        mOriginalTooltip = text;
        super.setToolTipText(text);
    }

    public String getErrorMessage() {
        return mErrorMsg;
    }

    public void setErrorMessage(String msg) {
        if (!Objects.equals(mErrorMsg, msg)) {
            Color foregroundColor;
            Color backgroundColor;
            mErrorMsg = msg;
            if (mErrorMsg == null) {
                foregroundColor = Colors.ON_EDITABLE;
                backgroundColor = Colors.EDITABLE;
                super.setToolTipText(mOriginalTooltip);
            } else {
                foregroundColor = Colors.ON_ERROR;
                backgroundColor = Colors.ERROR;
                if (mOriginalTooltip == null) {
                    super.setToolTipText(msg);
                } else {
                    super.setToolTipText(msg + "\n\n" + mOriginalTooltip);
                }
            }
            setForeground(foregroundColor);
            setBackground(backgroundColor);
            setCaretColor(foregroundColor);
            repaint();
        }
    }
}
