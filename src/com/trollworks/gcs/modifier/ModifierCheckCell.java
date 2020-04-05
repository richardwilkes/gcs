/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.MultipleRowUndo;
import com.trollworks.gcs.widgets.outline.RowUndo;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.widget.outline.Column;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.TextCell;

import java.awt.Color;
import java.awt.Component;
import java.awt.Font;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

public class ModifierCheckCell extends TextCell {
    private boolean mForEditor;

    public ModifierCheckCell(boolean forEditor) {
        super(SwingConstants.CENTER, false);
        mForEditor = forEditor;
    }

    @Override
    public Font getFont(Row row, Column column) {
        return mForEditor ? UIManager.getFont(GCSFonts.KEY_FIELD) : super.getFont(row, column);
    }

    @Override
    public Color getColor(boolean selected, boolean active, Row row, Column column) {
        if (mForEditor && row instanceof ListRow && !((ListRow) row).isSatisfied()) {
            return Color.red;
        }
        return super.getColor(selected, active, row, column);
    }

    @Override
    public String getToolTipText(Outline outline, MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (!mForEditor || !(row instanceof ListRow) || ((ListRow) row).isSatisfied()) {
            return super.getToolTipText(outline, event, bounds, row, column);
        }
        return ((ListRow) row).getReasonForUnsatisfied();
    }

    @SuppressWarnings("unused")
    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (row instanceof Modifier) {
            Modifier modifier = (Modifier) row;
            if (mForEditor) {
                modifier.setEnabled(!modifier.isEnabled());
                ModifierListEditor editor = UIUtilities.getAncestorOfType((Component) (event.getSource()), ModifierListEditor.class);
                if (editor != null) {
                    editor.mModified = true;
                }
            } else {
                RowUndo undo = new RowUndo(modifier);
                modifier.setEnabled(!modifier.isEnabled());
                if (undo.finish()) {
                    ArrayList<RowUndo> list = new ArrayList<>();
                    list.add(undo);
                    new MultipleRowUndo(list);
                }
            }
        }
    }
}
