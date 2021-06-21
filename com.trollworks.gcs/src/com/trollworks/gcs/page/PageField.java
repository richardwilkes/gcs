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

package com.trollworks.gcs.page;

import com.trollworks.gcs.character.CharacterSetter;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.DynamicColor;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.ToolTip;
import com.trollworks.gcs.utility.Platform;

import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.text.ParseException;
import javax.swing.JFormattedTextField;
import javax.swing.plaf.basic.BasicTextFieldUI;

/** A generic field for a page. */
public class PageField extends JFormattedTextField implements PropertyChangeListener, ActionListener, Commitable {
    private CharacterSheet  mSheet;
    private String          mTag;
    private CharacterSetter mSetter;
    private ThemeFont       mThemeFont;

    /**
     * Creates a new disabled text input field.
     *
     * @param factory      The {@link AbstractFormatterFactory} to use.
     * @param currentValue The current value.
     * @param sheet        The sheet the data belongs to.
     * @param alignment    The alignment of the field.
     * @param tooltip      The tooltip to set.
     */
    public PageField(AbstractFormatterFactory factory, Object currentValue, CharacterSheet sheet, int alignment, String tooltip) {
        this(factory, currentValue, null, sheet, "", alignment, false, tooltip);
    }

    /**
     * Creates a new text input field.
     *
     * @param factory      The {@link AbstractFormatterFactory} to use.
     * @param currentValue The current value.
     * @param sheet        The sheet the data belongs to.
     * @param tag          The tag for this field.
     * @param alignment    The alignment of the field.
     * @param editable     Whether or not the user can edit this field.
     * @param tooltip      The tooltip to set.
     */
    public PageField(AbstractFormatterFactory factory, Object currentValue, CharacterSetter setter, CharacterSheet sheet, String tag, int alignment, boolean editable, String tooltip) {
        super(factory, currentValue);
        if (Platform.isLinux()) {
            // I override the UI here since the GTK UI on Linux has no way to turn off the border
            // around text fields.
            setUI(new BasicTextFieldUI());
        }
        mSheet = sheet;
        mTag = tag;
        mSetter = setter;
        setThemeFont(ThemeFont.PAGE_FIELD_PRIMARY);
        setBorder(null);
        setOpaque(true);
        setHorizontalAlignment(alignment);
        setEditable(editable);
        setEnabled(editable);
        setForeground(editable ? ThemeColor.ON_EDITABLE : ThemeColor.ON_CONTENT);
        setBackground(editable ? ThemeColor.EDITABLE : ThemeColor.CONTENT);
        setSelectionColor(ThemeColor.SELECTION);
        setSelectedTextColor(ThemeColor.ON_SELECTION);
        setDisabledTextColor(new DynamicColor(() -> Colors.getWithAlpha(getForeground(), 128).getRGB()));
        setToolTipText(tooltip);
        addPropertyChangeListener("value", this);
        addActionListener(this);
        setFocusLostBehavior(COMMIT_OR_REVERT);
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
        return mSheet.getScale().scale(mThemeFont.getFont());
    }

    @Override
    public Dimension getPreferredSize() {
        Dimension size = super.getPreferredSize();
        // Don't know why this is needed, but it seems to be. Without it, text is being truncated by
        // about 2 pixels.
        size.width += mSheet.getScale().scale(2);
        return size;
    }

    @Override
    protected void processFocusEvent(FocusEvent event) {
        super.processFocusEvent(event);
        if (event.getID() == FocusEvent.FOCUS_GAINED) {
            selectAll();
        }
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        super.paintComponent(gc);
        if (isEditable()) {
            Rectangle bounds = getBounds();
            bounds.x = 0;
            bounds.y = 0;
            gc.setColor(ThemeColor.DIVIDER);
            int height = mSheet.getScale().scale(1);
            gc.fillRect(bounds.x, bounds.y + bounds.height - height, bounds.width, height);
            gc.setColor(getForeground());
        }
    }

    public String getTag() {
        return mTag;
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        if (isEditable()) {
            mSetter.setValue(mSheet.getCharacter(), getValue());
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
}
