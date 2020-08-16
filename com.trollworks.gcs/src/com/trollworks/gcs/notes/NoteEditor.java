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

package com.trollworks.gcs.notes;

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.LinkedLabel;
import com.trollworks.gcs.ui.widget.Workspace;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Dimension;
import javax.swing.JLabel;
import javax.swing.JScrollPane;
import javax.swing.JTextArea;
import javax.swing.JTextField;

/** The detailed editor for {@link Note}s. */
public class NoteEditor extends RowEditor<Note> {
    private JTextArea  mEditor;
    private JTextField mReferenceField;

    /**
     * Creates a new {@link Note} editor.
     *
     * @param note The {@link Note} to edit.
     */
    public NoteEditor(Note note) {
        super(note, new PrecisionLayout().setColumns(3).setMargins(0));
        add(new JLabel(note.getIcon(true)), new PrecisionLayoutData().setVerticalSpan(3).setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));

        add(new LinkedLabel(I18n.Text("Note Content")), new PrecisionLayoutData().setHorizontalSpan(2));
        mEditor = new JTextArea(note.getDescription());
        mEditor.setLineWrap(true);
        mEditor.setWrapStyleWord(true);
        mEditor.setEnabled(mIsEditable);
        JScrollPane scroller = new JScrollPane(mEditor);
        Dimension   size     = Workspace.get().getSize();
        size.width -= 200;
        if (size.width > 750) {
            size.width = 750;
        }
        size.height -= 200;
        if (size.height > 750) {
            size.height = 750;
        }
        if (size.width > 300 && size.height > 200) {
            scroller.setPreferredSize(size);
        }
        add(scroller, new PrecisionLayoutData().setHorizontalSpan(2).setFillAlignment().setGrabSpace(true).setMinimumWidth(300).setMinimumHeight(200));

        add(new LinkedLabel(I18n.Text("Page Reference"), mReferenceField));
        mReferenceField = new JTextField(Text.makeFiller(6, 'M'));
        mReferenceField.setText(note.getReference());
        mReferenceField.setToolTipText(Text.wrapPlainTextForToolTip(PageRefCell.getStdToolTip(I18n.Text("note"))));
        mReferenceField.setEnabled(mIsEditable);
        add(mReferenceField, new PrecisionLayoutData().setGrabHorizontalSpace(true).setFillHorizontalAlignment());
    }

    @Override
    public boolean applyChangesSelf() {
        boolean modified = mRow.setReference(mReferenceField.getText());
        modified |= mRow.setDescription(mEditor.getText());
        return modified;
    }

    @Override
    public void finished() {
        // Unused
    }
}
