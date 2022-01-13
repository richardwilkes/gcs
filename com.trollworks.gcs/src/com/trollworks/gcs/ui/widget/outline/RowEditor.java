/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.ActionPanel;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.ui.widget.LayoutConstants;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.MultiLineTextField;
import com.trollworks.gcs.ui.widget.ScrollContent;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Rectangle;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JComponent;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentListener;

/**
 * The base class for all row editors.
 *
 * @param <T> The row class being edited.
 */
public abstract class RowEditor<T extends ListRow> extends ActionPanel {
    private static final int     MARGIN = 10;
    /** Whether the underlying data should be editable. */
    protected            boolean mIsEditable;
    /** The row being edited. */
    protected            T       mRow;

    /**
     * Brings up a modal detailed editor for each row in the list.
     *
     * @param owner The owning component.
     * @param list  The rows to edit.
     * @return Whether anything was modified.
     */
    public static boolean edit(Component owner, List<? extends ListRow> list) {
        ArrayList<RowUndo> undos  = new ArrayList<>();
        ListRow[]          rows   = list.toArray(new ListRow[0]);
        int                length = rows.length;
        for (int i = 0; i < length; i++) {
            ListRow   row     = rows[i];
            Modal     dialog  = new Modal(owner, MessageFormat.format(I18n.text("Edit {0}"), row.getRowType()));
            Container content = dialog.getContentPane();
            if (i != length - 1) {
                int remaining = length - i - 1;
                String msg = remaining == 1 ? I18n.text("1 item remaining to be edited.") :
                        String.format(I18n.text("%s items remaining to be edited."), Numbers.format(remaining));
                Label label = new Label(msg, SwingConstants.CENTER);
                label.setBorder(new EmptyBorder(LayoutConstants.WINDOW_BORDER_INSET, 0, 0, 0));
                content.add(label, BorderLayout.NORTH);
                dialog.addCancelRemainingButton();
            }
            RowEditor<? extends ListRow> editor = row.createEditor();
            content.add(editor, BorderLayout.CENTER);
            dialog.addCancelButton();
            dialog.addApplyButton();
            dialog.presentToUser();
            switch (dialog.getResult()) {
                case Modal.OK -> {
                    RowUndo undo = new RowUndo(row);
                    if (editor.applyChanges()) {
                        if (undo.finish()) {
                            undos.add(undo);
                        }
                    }
                }
                case Modal.CLOSED -> i = length;
            }
        }
        if (!undos.isEmpty()) {
            new MultipleRowUndo(undos);
            return true;
        }
        return false;
    }

    /**
     * Creates a new RowEditor.
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
        ScrollContent outer = new ScrollContent(new PrecisionLayout().
                setMargins(LayoutConstants.WINDOW_BORDER_INSET, LayoutConstants.WINDOW_BORDER_INSET,
                        0, LayoutConstants.WINDOW_BORDER_INSET));
        outer.setScrollableTracksViewportWidth(true);
        outer.setBackground(Colors.BACKGROUND);
        addContentSelf(outer);
        if (!mIsEditable) {
            UIUtilities.disableControls(outer);
        }
        ScrollPanel scroller = new ScrollPanel(outer);
        scroller.setPreferredSize(adjustSize(outer.getPreferredSize()));
        Dimension size = adjustSize(outer.getMinimumSize());
        if (size.height > 128) {
            size.height = 128;
        }
        scroller.setMinimumSize(size);
        add(scroller);
    }

    private static Dimension adjustSize(Dimension size) {
        size = new Dimension(size);
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

    protected static void addSection(Container parent, JComponent section) {
        section.setBorder(new TitledBorder(Fonts.HEADER, section.toString()));
        parent.add(section, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));
    }

    protected static PrecisionLayoutData addLabel(Container parent, String text) {
        PrecisionLayoutData layoutData = new PrecisionLayoutData();
        parent.add(new Label(text), layoutData.setEndHorizontalAlignment());
        return layoutData;
    }

    protected static void addLabelPinnedToTop(Container parent, String text) {
        parent.add(new Label(text), new PrecisionLayoutData().setEndHorizontalAlignment().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING).setTopMargin(2));
    }

    protected static PrecisionLayoutData addInteriorLabel(Container parent, String text) {
        PrecisionLayoutData layoutData = new PrecisionLayoutData();
        parent.add(new Label(text), layoutData.setEndHorizontalAlignment().setLeftMargin(4));
        return layoutData;
    }

    public MultiLineTextField addVTTNotesField(Container parent, DocumentListener listener) {
        return addVTTNotesField(parent, 1, listener);
    }

    public MultiLineTextField addVTTNotesField(Container parent, int hSpan, DocumentListener listener) {
        MultiLineTextField field = new MultiLineTextField(mRow.getVTTNotes(),
                I18n.text("Any notes for VTT use; see the instructions for your VTT to determine if/how these can be used"), listener);
        addLabel(parent, I18n.text("VTT Notes")).setBeginningVerticalAlignment().setTopMargin(2);
        parent.add(field, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true).setHorizontalSpan(hSpan));
        return field;
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
