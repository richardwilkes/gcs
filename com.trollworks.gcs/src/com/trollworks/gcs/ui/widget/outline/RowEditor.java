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

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.ActionPanel;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;

/**
 * The base class for all row editors.
 *
 * @param <T> The row class being edited.
 */
public abstract class RowEditor<T extends ListRow> extends ActionPanel {
    /** Whether the underlying data should be editable. */
    protected boolean mIsEditable;
    /** The row being edited. */
    protected T       mRow;

    /**
     * Brings up a modal detailed editor for each row in the list.
     *
     * @param owner The owning component.
     * @param list  The rows to edit.
     * @return Whether anything was modified.
     */
    public static boolean edit(Component owner, List<? extends ListRow> list) {
        ArrayList<RowUndo> undos = new ArrayList<>();
        ListRow[]          rows  = list.toArray(new ListRow[0]);

        int length = rows.length;
        for (int i = 0; i < length; i++) {
            boolean                      hasMore = i != length - 1;
            ListRow                      row     = rows[i];
            RowEditor<? extends ListRow> editor  = row.createEditor();
            String                       title   = MessageFormat.format(I18n.Text("Edit {0}"), row.getRowType());
            JPanel                       wrapper = new JPanel(new BorderLayout());

            if (hasMore) {
                int    remaining = length - i - 1;
                String msg       = remaining == 1 ? I18n.Text("1 item remaining to be edited.") : MessageFormat.format(I18n.Text("{0} items remaining to be edited."), Integer.valueOf(remaining));
                JLabel panel     = new JLabel(msg, SwingConstants.CENTER);
                panel.setBorder(new EmptyBorder(0, 0, 10, 0));
                wrapper.add(panel, BorderLayout.NORTH);
            }
            wrapper.add(editor, BorderLayout.CENTER);

            int      type       = hasMore ? JOptionPane.YES_NO_CANCEL_OPTION : JOptionPane.YES_NO_OPTION;
            String   applyText  = I18n.Text("Apply");
            String   cancelText = I18n.Text("Cancel");
            String[] options    = hasMore ? new String[]{applyText, cancelText, I18n.Text("Cancel Remaining")} : new String[]{applyText, cancelText};
            switch (WindowUtils.showOptionDialog(owner, wrapper, title, true, type, JOptionPane.PLAIN_MESSAGE, null, options, applyText)) {
            case JOptionPane.YES_OPTION:
                RowUndo undo = new RowUndo(row);
                if (editor.applyChanges()) {
                    if (undo.finish()) {
                        undos.add(undo);
                    }
                }
                break;
            case JOptionPane.NO_OPTION:
                break;
            case JOptionPane.CANCEL_OPTION:
            case JOptionPane.CLOSED_OPTION:
            default:
                i = length;
                break;
            }
        }

        if (!undos.isEmpty()) {
            new MultipleRowUndo(undos);
            return true;
        }
        return false;
    }

    /**
     * Creates a new {@link RowEditor}.
     *
     * @param row The row being edited.
     */
    protected RowEditor(T row) {
        super(new BorderLayout());
        mRow = row;
        mIsEditable = !mRow.getOwner().isLocked();
    }

    /** Call this during construction to add the content to the editor. */
    protected void addContent() {
        ScrollContent outer = new ScrollContent(new PrecisionLayout().setMargins(0));
        addContentSelf(outer);
        if (!mIsEditable) {
            UIUtilities.disableControls(outer);
        }
        outer.setScrollableTracksViewportWidth(true);
        JScrollPane scroller   = new JScrollPane(outer);
        int         scrollSize = scroller.getVerticalScrollBar().getPreferredSize().width;
        scroller.setPreferredSize(adjustSize(outer.getPreferredSize(), scrollSize));
        Dimension size = adjustSize(outer.getMinimumSize(), scrollSize);
        if (size.height > 128) {
            size.height = 128;
        }
        scroller.setMinimumSize(size);
        scroller.setBorder(null);
        scroller.getViewport().setBackground(outer.getBackground());
        add(scroller);
    }

    private Dimension adjustSize(Dimension size, int scrollSize) {
        size = new Dimension(size);
        size.width += scrollSize;
        size.height += scrollSize;
        Rectangle maxBounds = WindowUtils.getMaximumWindowBounds();
        if (size.width > maxBounds.width - 64) {
            size.width = maxBounds.width - 64;
        }
        if (size.height > maxBounds.height - 64) {
            size.height = maxBounds.height - 64;
        }
        return size;
    }

    /** Called by {@link #addContent()}. */
    protected abstract void addContentSelf(ScrollContent outer);

    protected final void addSection(Container parent, JComponent section) {
        section.setBorder(new TitledBorder(Fonts.getDefaultSystemFont(), section.toString()));
        parent.add(section, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    protected final void addLabel(Container parent, String text, JComponent linkedTo) {
        parent.add(new LinkedLabel(text, linkedTo), new PrecisionLayoutData().setFillHorizontalAlignment());
    }

    /**
     * Called to apply any changes that were made.
     *
     * @return Whether anything was modified.
     */
    public final boolean applyChanges() {
        Commitable.sendCommitToFocusOwner();
        boolean modified = applyChangesSelf();
        if (modified) {
            mRow.getDataFile().setModified(true);
        }
        return modified;
    }

    /**
     * Called to apply any changes that were made.
     *
     * @return Whether anything was modified.
     */
    protected abstract boolean applyChangesSelf();
}
