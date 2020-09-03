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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.MultipleRowUndo;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowUndo;
import com.trollworks.gcs.ui.widget.outline.TextCell;

import java.awt.Color;
import java.awt.Component;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import javax.swing.SwingConstants;

public class ModifierCheckCell extends TextCell {
    private boolean mForEditor;

    public ModifierCheckCell(boolean forEditor) {
        super(SwingConstants.CENTER, false);
        mForEditor = forEditor;
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

    @Override
    public void mouseClicked(MouseEvent event, Rectangle bounds, Row row, Column column) {
        if (row instanceof Modifier) {
            Component src      = (Component) event.getSource();
            Modifier  modifier = (Modifier) row;
            if (mForEditor) {
                modifier.setEnabled(!modifier.isEnabled());
                ModifierListEditor editor = UIUtilities.getAncestorOfType(src, ModifierListEditor.class);
                if (editor != null) {
                    editor.mModified = true;
                    editor.notifyActionListeners();
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
            Outline outline = UIUtilities.getSelfOrAncestorOfType(src, Outline.class);
            if (outline != null) {
                outline.repaintSelection();
            }
            event.consume();
        }
    }
}
