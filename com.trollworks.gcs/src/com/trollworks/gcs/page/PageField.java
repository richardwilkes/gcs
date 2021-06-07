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
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.text.ParseException;
import javax.swing.JFormattedTextField;
import javax.swing.UIManager;
import javax.swing.plaf.basic.BasicTextFieldUI;

/** A generic field for a page. */
public class PageField extends JFormattedTextField implements PropertyChangeListener, ActionListener, Commitable {
    private CharacterSheet  mSheet;
    private String          mTag;
    private CharacterSetter mSetter;

    /**
     * Creates a new disabled text input field.
     *
     * @param factory      The {@link AbstractFormatterFactory} to use.
     * @param currentValue The current value.
     * @param sheet        The sheet the data belongs to.
     * @param alignment    The alignment of the field.
     * @param tooltip      The tooltip to set.
     * @param color        The color to use.
     */
    public PageField(AbstractFormatterFactory factory, Object currentValue, CharacterSheet sheet, int alignment, String tooltip, Color color) {
        this(factory, currentValue, null, sheet, "", alignment, false, tooltip, color);
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
     * @param color        The color to use.
     */
    public PageField(AbstractFormatterFactory factory, Object currentValue, CharacterSetter setter, CharacterSheet sheet, String tag, int alignment, boolean editable, String tooltip, Color color) {
        super(factory, currentValue);
        if (Platform.isLinux()) {
            // I override the UI here since the GTK UI on Linux has no way to turn off the border
            // around text fields.
            setUI(new BasicTextFieldUI());
        }
        mSheet = sheet;
        mTag = tag;
        mSetter = setter;
        setFont(sheet.getScale().scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY)));
        setBorder(null);
        setOpaque(false);
        // Just setting opaque to false isn't enough for some reason, so I'm also setting the
        // background color to a 100% transparent value.
        setBackground(Colors.TRANSPARENT);
        setHorizontalAlignment(alignment);
        setEditable(editable);
        setEnabled(editable);
        setForeground(editable ? ThemeColor.ON_EDITABLE : color);
        setDisabledTextColor(color);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        addPropertyChangeListener("value", this);
        addActionListener(this);
        setFocusLostBehavior(COMMIT_OR_REVERT);
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
    protected void paintComponent(Graphics gc) {
        super.paintComponent(GraphicsUtilities.prepare(gc));
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
}
